package cbrotli_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/CAFxX/httpcompression"
	"github.com/CAFxX/httpcompression/contrib/google/cbrotli"
	gcbrotli "github.com/google/brotli/go/cbrotli"
)

var _ httpcompression.CompressorProvider = &cbrotli.Compressor{}

func TestZstd(t *testing.T) {
	t.Parallel()

	s := []byte("hello world!")

	c, err := cbrotli.New(gcbrotli.WriterOptions{})
	if err != nil {
		t.Fatal(err)
	}
	b := &bytes.Buffer{}
	w := c.Get(b)
	w.Write(s)
	w.Close()

	r := gcbrotli.NewReader(b)
	if err != nil {
		t.Fatal(err)
	}
	d, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(s, d) {
		t.Fatalf("decoded string mismatch\ngot: %q\nexp: %q", string(s), string(d))
	}
}
