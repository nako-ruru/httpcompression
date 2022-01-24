package zlib

import (
	"fmt"
	"io"
	"sync"

	"github.com/CAFxX/httpcompression/contrib/internal/utils"
	"github.com/klauspost/compress/zlib"
)

const (
	Encoding           = "gzip"
	DefaultCompression = zlib.DefaultCompression
)

type compressor struct {
	pool sync.Pool
	opts Options
}

type Options struct {
	Level      int
	Dictionary []byte
}

func New(opts Options) (c *compressor, err error) {
	defer func() {
		if r := recover(); r != nil {
			c, err = nil, fmt.Errorf("panic: %v", r)
		}
	}()

	tw, err := zlib.NewWriterLevelDict(io.Discard, opts.Level, opts.Dictionary)
	if err != nil {
		return nil, err
	}
	if err := utils.CheckWriter(tw); err != nil {
		return nil, fmt.Errorf("deflate: writer initialization: %w", err)
	}

	c = &compressor{opts: opts}
	return c, nil
}

func (c *compressor) Get(w io.Writer) io.WriteCloser {
	if gw, ok := c.pool.Get().(*writer); ok {
		gw.Reset(w)
		return gw
	}
	gw, err := zlib.NewWriterLevelDict(w, c.opts.Level, c.opts.Dictionary)
	if err != nil {
		return utils.ErrorWriteCloser{Err: err}
	}
	return &writer{
		Writer: gw,
		c:      c,
	}
}

type writer struct {
	*zlib.Writer
	c *compressor
}

func (w *writer) Close() error {
	err := w.Writer.Close()
	w.Reset(nil)
	w.c.pool.Put(w)
	return err
}
