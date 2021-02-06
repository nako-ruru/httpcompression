package zstd

import (
	"io"
	"runtime"
	"sync"

	"github.com/valyala/gozstd"
)

const (
	Encoding           = "zstd"
	DefaultCompression = gozstd.DefaultCompressionLevel
)

type WriterParams = gozstd.WriterParams

type compressor struct {
	pool sync.Pool
	opts WriterParams
}

func New(opts *WriterParams) (c *compressor, err error) {
	c = &compressor{}
	if opts != nil {
		c.opts = *opts // make a copy of opts to ensure its contents do not change
	}
	gw := gozstd.NewWriterParams(nil, &c.opts)
	c.pool.Put(gw)
	return c, nil
}

func (c *compressor) Get(w io.Writer) io.WriteCloser {
	if gw, ok := c.pool.Get().(*writer); ok {
		gw.ResetWriterParams(w, &c.opts)
		return gw
	}
	cw := gozstd.NewWriterParams(w, &c.opts)
	gw := &writer{
		Writer: cw,
		c:      c,
	}
	runtime.SetFinalizer(gw, func(gw *writer) {
		// The gozstd implementation internally does the same thing, but since
		// that behavior is not guaranteed by their API, we do the same here.
		gw.Writer.Release()
	})
	return gw
}

type writer struct {
	*gozstd.Writer
	c *compressor
}

func (w *writer) Close() error {
	err := w.Writer.Close()
	w.Writer.ResetWriterParams(nil, &w.c.opts)
	w.c.pool.Put(w)
	return err
}
