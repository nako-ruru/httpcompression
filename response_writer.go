package httpcompression

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
)

// 0 < minSize <= maxCache <= bufSize

// compressWriter provides an http.ResponseWriter interface, which gzips
// bytes before writing them to the underlying response. This doesn't close the
// writers, so don't forget to do that.
// It can be configured to skip response smaller than minSize.
type compressWriter struct {
	http.ResponseWriter

	config config
	accept codings
	common []string

	w     io.Writer
	enc   string
	code  int // Saves the WriteHeader value.
	chunk []byte
	err   error
}

var (
	_ io.WriteCloser = &compressWriter{}
	_ http.Flusher   = &compressWriter{}
	_ http.Hijacker  = &compressWriter{}
	_ writeStringer  = &compressWriter{}
)

type compressWriterWithCloseNotify struct {
	*compressWriter
}

func (w compressWriterWithCloseNotify) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

var (
	_ io.WriteCloser = compressWriterWithCloseNotify{}
	_ http.Flusher   = compressWriterWithCloseNotify{}
	_ http.Hijacker  = compressWriterWithCloseNotify{}
	_ writeStringer  = compressWriterWithCloseNotify{}
)

// Write compresses and appends the given byte slice to the underlying ResponseWriter.
func (w *compressWriter) Write(b []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}

	if cap(w.chunk) == 0 || (len(w.chunk) == 0 && len(b) >= cap(w.chunk)) {
		return w.writeChunk(b, false)
	}

	max := cap(w.chunk) - len(w.chunk)
	if max >= len(b) {
		w.chunk = append(w.chunk, b...)
		return len(b), nil
	}

	w.chunk = append(w.chunk, b[:max]...)
	_, err := w.writeChunk(w.chunk, false)
	if err != nil {
		return 0, err
	}
	w.chunk = w.chunk[:0]

	if len(b[max:]) < cap(w.chunk) {
		w.chunk = append(w.chunk, b[max:]...)
		return len(b), nil
	}
	_, err = w.writeChunk(b[max:], false)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

func (w *compressWriter) writeChunk(b []byte, final bool) (int, error) {
	if w.w != nil {
		// The responseWriter is already initialized: use it.
		return w.w.Write(b)
	}

	if clv := w.Header().Get(contentLength); clv != "" {
		cl, _ := strconv.Atoi(clv)
		if cl < w.config.minSize {
			goto plain
		}
	}
	if final && (len(b) < w.config.minSize || len(b) == 0) {
		goto plain
	}

	if w.Header().Get(contentEncoding) != "" {
		goto plain
	}

	{
		// If a Content-Type wasn't specified, infer it from the current buffer.
		ct := w.Header().Get(contentType)
		if ct == "" {
			ct = http.DetectContentType(b)
			if ct != "" {
				// net/http by default performs content sniffing but this is disabled if content-encoding is set.
				// Since we set content-encoding, if content-type was not set and we successfully sniffed it,
				// set the content-type.
				w.Header().Set(contentType, ct)
			}
		}

		if handleContentType(ct, w.config.contentTypes, w.config.blacklist) == false {
			goto plain
		}

		enc := preferredEncoding(w.accept, w.config.compressor, w.common, w.config.prefer)
		if err := w.startCompress(b, enc); err != nil {
			return 0, err
		}
		return len(b), nil
	}

plain:
	// If we got here, we should not GZIP this response.
	if err := w.startPlain(b); err != nil {
		return 0, err
	}
	return len(b), nil
}

// WriteString compresses and appends the given string to the underlying ResponseWriter.
//
// This makes use of an optional method (WriteString) exposed by the compressors, or by
// the underlying ResponseWriter.
func (w *compressWriter) WriteString(s string) (int, error) {
	if w.err != nil {
		return 0, w.err
	}

	// Since WriteString is an optional interface of the compressor, and the actual compressor
	// is chosen only after the first call to Write, we can't statically know whether the interface
	// is supported. We therefore have to check dynamically.
	if ws, _ := w.w.(writeStringer); ws != nil && len(w.chunk) == 0 && len(s) > cap(w.chunk) {
		// The responseWriter is already initialized and it implements WriteString.
		return ws.WriteString(s)
	}
	// Fallback: the writer has not been initialized yet, or it has been initialized
	// and it does not implement WriteString. We could in theory do something unsafe
	// here but for now let's keep it simple and fallback to Write.
	return w.Write([]byte(s))
}

type writeStringer interface {
	WriteString(string) (int, error)
}

// startCompress initializes a compressing writer and writes the buffer.
func (w *compressWriter) startCompress(buf []byte, enc string) error {
	comp, ok := w.config.compressor[enc]
	if !ok {
		panic("unknown compressor")
	}

	w.Header().Set(contentEncoding, enc)

	// if the Content-Length is already set, then calls to Write on gzip
	// will fail to set the Content-Length header since its already set
	// See: https://github.com/golang/go/issues/14975.
	w.Header().Del(contentLength)

	// Write the header to gzip response.
	w.writeHeader()

	if len(buf) == 0 {
		panic("unreachable: empty buffer")
	}

	w.w = comp.comp.Get(w.ResponseWriter)
	w.enc = enc

	_, err := shortWriteChecker{w.w}.Write(buf)
	w.err = err
	return err
}

// startPlain writes to sent bytes and buffer the underlying ResponseWriter without gzip.
func (w *compressWriter) startPlain(buf []byte) error {
	w.writeHeader()
	w.w = w.ResponseWriter
	w.enc = ""
	// If Write was never called then don't call Write on the underlying ResponseWriter.
	if len(buf) == 0 {
		return nil
	}
	_, err := shortWriteChecker{w.ResponseWriter}.Write(buf)
	return err
}

// WriteHeader sets the response code that will be returned in the response.
func (w *compressWriter) WriteHeader(code int) {
	if w.code == 0 {
		w.code = code
	}
}

func (w *compressWriter) writeHeader() {
	if w.code != 0 {
		w.ResponseWriter.WriteHeader(w.code)
		// Ensure that no other WriteHeader's happen
		w.code = 0
	}
}

// Close closes the compression Writer.
func (w *compressWriter) Close() error {
	defer func() {
		w.err = errClosed
	}()

	if w.err != nil {
		return w.err
	}

	if len(w.chunk) > 0 {
		err := w.flushChunk(true)
		if err != nil {
			return err
		}
	}

	if w.w != nil && w.enc == "" {
		return nil
	}
	if cw, ok := w.w.(io.Closer); ok {
		w.w = nil
		return cw.Close()
	}

	// compression not triggered yet, write out regular response.
	err := w.startPlain(nil)
	// Returns the error if any at write.
	if err != nil {
		err = fmt.Errorf("httpcompression: write to regular responseWriter at close gets error: %v", err)
	}
	return err
}

var errClosed = fmt.Errorf("compressWriter already closed")

// Flush flushes the underlying compressor Writer and then the underlying
// http.ResponseWriter if it is an http.Flusher. This makes compressWriter
// an http.Flusher.
// Flush is a no-op until enough data has been written to decide whether the
// response should be compressed or not (e.g. less than MinSize bytes have
// been written).
func (w *compressWriter) Flush() {
	if w.err != nil {
		return
	}

	if err := w.flushChunk(false); err != nil {
		return
	}

	if w.w == nil {
		return
	}

	// Flush the compressor, if supported.
	// note: http.ResponseWriter does not implement Flusher (http.Flusher does not return an error),
	// so we need to later call ResponseWriter.Flush anyway:
	// - in case we are bypassing compression, w.w is the parent ResponseWriter, and therefore we skip
	//   this as the parent ResponseWriter does not implement Flusher.
	// - in case we are NOT bypassing compression, w.w is the compressor, and therefore we flush the
	//   compressor and then we flush the parent ResponseWriter.
	if fw, ok := w.w.(Flusher); ok {
		_ = fw.Flush()
	}

	// Flush the ResponseWriter (the previous Flusher is not expected to flush the parent writer).
	if fw, ok := w.ResponseWriter.(http.Flusher); ok {
		fw.Flush()
	}
}

func (w *compressWriter) flushChunk(final bool) error {
	if len(w.chunk) > 0 && (w.w != nil || final) {
		_, err := w.writeChunk(w.chunk, final)
		if err != nil {
			return err
		}
		w.chunk = w.chunk[:0]
	}
	return nil
}

// Hijack implements http.Hijacker. If the underlying ResponseWriter is a
// Hijacker, its Hijack method is returned. Otherwise an error is returned.
func (w *compressWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, fmt.Errorf("http.Hijacker interface is not supported")
}

type shortWriteChecker struct {
	io.Writer
}

func (w shortWriteChecker) Write(buf []byte) (int, error) {
	n, err := w.Writer.Write(buf)
	if err == nil && n < len(buf) {
		return 0, io.ErrShortWrite
	}
	return n, err
}
