package cbrotli

import (
	"fmt"
	"io"
	"runtime"

	"github.com/google/brotli/go/cbrotli"
)

const (
	Encoding = "br"
)

type compressor struct {
	opts cbrotli.WriterOptions
}

func New(opts cbrotli.WriterOptions) (c *compressor, err error) {
	tw := cbrotli.NewWriter(io.Discard, opts)
	n, err := tw.Write([]byte("test"))
	if n != 4 || err != nil {
		return nil, fmt.Errorf("cbrotli: writer initialization (write): %d, %w", n, err)
	}
	err = tw.Close()
	if err != nil {
		return nil, fmt.Errorf("cbrotli: writer initialization (close): %w", err)
	}

	c = &compressor{opts: opts}
	return c, nil
}

func (c *compressor) Get(w io.Writer) io.WriteCloser {
	gw := cbrotli.NewWriter(w, c.opts)

	cw := &writer{
		Writer: gw,
	}
	runtime.SetFinalizer(cw, func(cw *writer) {
		go func() {
			defer func() {
				recover()
			}()
			_ = cw.Close()
		}()
	})
	return cw
}

type writer struct {
	*cbrotli.Writer
}

func (w *writer) Write(buf []byte) (int, error) {
	return w.Writer.Write(buf)
}

func (w *writer) Close() error {
	defer runtime.SetFinalizer(w, nil)
	return w.Writer.Close()
}
