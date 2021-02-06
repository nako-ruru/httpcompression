package brotli

import (
	"fmt"
	"io"
	"runtime"

	"github.com/google/brotli/go/cbrotli"
)

const (
	Encoding = "br"
)

type Options = cbrotli.WriterOptions

type compressor struct {
	opts Options
}

func New(opts Options) (c *compressor, err error) {
	defer func() {
		if r := recover(); r != nil {
			c, err = nil, fmt.Errorf("panic: %v", r)
		}
	}()
	gw := cbrotli.NewWriter(nil, opts)
	gw.Close()
	c = &compressor{opts: opts}
	return c, nil
}

func (c *compressor) Get(w io.Writer) io.WriteCloser {
	// TODO: Add a pool once cbrotli supports Reset
	cw := cbrotli.NewWriter(w, c.opts)
	gw := &writer{
		Writer: cw,
	}
	runtime.SetFinalizer(gw, func(gw *writer) {
		gw.Close()
	})
	return gw
}

type writer struct {
	*cbrotli.Writer
}

func (gw *writer) Close() error {
	if gw.Writer == nil {
		return nil
	}
	err := gw.Writer.Close()
	gw.Writer = nil
	return err
}
