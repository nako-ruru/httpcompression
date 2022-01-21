package pgzip

import (
	"fmt"
	"io"
	"sync"

	"github.com/CAFxX/httpcompression/contrib/internal/utils"
	"github.com/klauspost/pgzip"
)

const (
	Encoding           = "gzip"
	DefaultCompression = pgzip.DefaultCompression
)

type compressor struct {
	pool sync.Pool
	opts Options
}

type Options struct {
	Level     int
	BlockSize int
	Blocks    int
}

func New(opts Options) (c *compressor, err error) {
	defer func() {
		if r := recover(); r != nil {
			c, err = nil, fmt.Errorf("panic: %v", r)
		}
	}()

	tw, err := pgzip.NewWriterLevel(io.Discard, opts.Level)
	if err != nil {
		return nil, err
	}
	err = tw.SetConcurrency(opts.BlockSize, opts.Blocks)
	if err != nil {
		return nil, err
	}
	if err := utils.CheckWriter(tw); err != nil {
		return nil, fmt.Errorf("pgzip: writer initialization: %w", err)
	}

	c = &compressor{opts: opts}
	return c, nil
}

func (c *compressor) Get(w io.Writer) io.WriteCloser {
	if gw, ok := c.pool.Get().(*writer); ok {
		gw.Reset(w)
		return gw
	}
	gw, err := pgzip.NewWriterLevel(w, c.opts.Level)
	if err != nil {
		return utils.ErrorWriteCloser{Err: err}
	}
	err = gw.SetConcurrency(c.opts.BlockSize, c.opts.Blocks)
	if err != nil {
		return utils.ErrorWriteCloser{Err: err}
	}
	return &writer{
		Writer: gw,
		c:      c,
	}
}

type writer struct {
	*pgzip.Writer
	c *compressor
}

func (w *writer) Close() error {
	err := w.Writer.Close()
	w.Reset(nil)
	w.c.pool.Put(w)
	return err
}
