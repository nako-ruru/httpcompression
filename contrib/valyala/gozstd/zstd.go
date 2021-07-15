package gozstd

import (
	"io"
	"sync"

	"github.com/valyala/gozstd"
)

const (
	Encoding = "zstd"
)

type compressor struct {
	pool sync.Pool
	opts gozstd.WriterParams
}

func New(opts gozstd.WriterParams) (c *compressor, err error) {
	c = &compressor{opts: opts}
	return c, nil
}

func (c *compressor) Get(w io.Writer) io.WriteCloser {
	if gw, ok := c.pool.Get().(*zstdWriter); ok {
		gw.ResetWriterParams(w, &c.opts)
		return gw
	}
	gw := gozstd.NewWriterParams(w, &c.opts)
	return &zstdWriter{
		Writer: gw,
		c:      c,
	}
}

type zstdWriter struct {
	*gozstd.Writer
	c *compressor
}

func (w *zstdWriter) Close() error {
	err := w.Writer.Close()
	w.ResetWriterParams(nil, &w.c.opts) // drop reference to parent writer
	w.c.pool.Put(w)
	return err
}
