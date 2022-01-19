package cbrotli

import (
	"fmt"
	"io"
	"runtime"

	"github.com/CAFxX/httpcompression/contrib/internal/utils"
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
	if err := utils.CheckWriter(tw); err != nil {
		return nil, fmt.Errorf("cbrotli: writer initialization: %w", err)
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
