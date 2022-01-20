package flate

import (
	"compress/flate"
	"io"
	"sync"

	"github.com/CAFxX/httpcompression/contrib/internal/utils"
)

type Options struct {
	Level      int
	Dictionary []byte
}

type compressor struct {
	pool sync.Pool
	opt  Options
}

func New(opt Options) (*compressor, error) {
	tw, err := flate.NewWriterDict(io.Discard, opt.Level, opt.Dictionary)
	if err != nil {
		return nil, err
	}
	err = utils.CheckWriter(tw)
	if err != nil {
		return nil, err
	}

	c := &compressor{opt: opt}
	return c, nil
}

func (c *compressor) Get(w io.Writer) io.WriteCloser {
	if gw, ok := c.pool.Get().(*deflateWriter); ok {
		gw.Reset(w)
		return gw
	}
	gw, _ := flate.NewWriterDict(w, c.opt.Level, c.opt.Dictionary)
	return &deflateWriter{
		Writer: gw,
		c:      c,
	}
}

type deflateWriter struct {
	*flate.Writer
	c *compressor
}

func (w *deflateWriter) Close() error {
	err := w.Writer.Close()
	w.Reset(nil)
	w.c.pool.Put(w)
	return err
}
