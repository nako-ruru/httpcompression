package lz4

import (
	"fmt"
	"io"
	"sync"

	"github.com/CAFxX/httpcompression/contrib/internal/utils"
	lz4 "github.com/pierrec/lz4/v4"
)

const (
	Encoding = "lz4"
)

type compressor struct {
	pool sync.Pool
	opts []lz4.Option
}

func New(opts ...lz4.Option) (c *compressor, err error) {
	defer func() {
		if r := recover(); r != nil {
			c, err = nil, fmt.Errorf("panic: %v", r)
		}
	}()

	opts = append([]lz4.Option(nil), opts...)

	tw := lz4.NewWriter(io.Discard)
	err = tw.Apply(opts...)
	if err != nil {
		return nil, fmt.Errorf("lz4: apply options: %w", err)
	}
	if err := utils.CheckWriter(tw); err != nil {
		return nil, fmt.Errorf("lz4: writer initialization: %w", err)
	}

	c = &compressor{opts: opts}
	return c, nil
}

func (c *compressor) Get(w io.Writer) io.WriteCloser {
	if gw, ok := c.pool.Get().(*writer); ok {
		gw.Reset(w)
		return gw
	}
	gw := lz4.NewWriter(w)
	err := gw.Apply(c.opts...)
	if err != nil {
		return utils.ErrorWriteCloser{Err: err}
	}
	return &writer{
		Writer: gw,
		c:      c,
	}
}

type writer struct {
	*lz4.Writer
	c *compressor
}

func (w *writer) Close() error {
	err := w.Writer.Close()
	w.Reset(nil)
	w.c.pool.Put(w)
	return err
}
