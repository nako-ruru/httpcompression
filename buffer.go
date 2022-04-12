package httpcompression

import (
	"bufio"
	"io"
	"sync"
)

type bufferedCompressorProvider struct {
	c CompressorProvider
	p *sync.Pool
}

func withBuffer(n int) func(CompressorProvider) CompressorProvider {
	p := &sync.Pool{}
	p.New = func() interface{} {
		c := new(bufferedCompressor)
		c.Writer = *bufio.NewWriterSize(nil, n)
		c.pool = p
		return c
	}
	return func(c CompressorProvider) CompressorProvider {
		return &bufferedCompressorProvider{c: c, p: p}
	}
}

func (p *bufferedCompressorProvider) Get(w io.Writer) io.WriteCloser {
	c := p.c.Get(w)

	bc := p.p.Get().(*bufferedCompressor)
	bc.Reset(c)
	bc.parent = c

	return bc
}

type bufferedCompressor struct {
	bufio.Writer
	parent io.WriteCloser
	pool   *sync.Pool
}

func (c *bufferedCompressor) Flush() error {
	err := c.Writer.Flush()
	if err != nil {
		return err
	}
	if fp, ok := c.parent.(Flusher); ok {
		err = fp.Flush()
	}
	return err
}

func (c *bufferedCompressor) Close() error {
	err := c.Writer.Flush()
	errc := c.parent.Close()
	if err == nil {
		err = errc
	}

	c.Reset(nil)
	c.parent = nil
	c.pool.Put(c)

	return err
}
