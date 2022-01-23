package lz4_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/CAFxX/httpcompression"
	"github.com/CAFxX/httpcompression/contrib/pierrec/lz4"
	plz4 "github.com/pierrec/lz4/v4"
)

var _ httpcompression.CompressorProvider = &lz4.Compressor{}

func TestLz4(t *testing.T) {
	t.Parallel()

	s := []byte("hello world!")

	c, err := lz4.New()
	if err != nil {
		t.Fatal(err)
	}
	b := &bytes.Buffer{}
	w := c.Get(b)
	w.Write(s)
	w.Close()

	r := plz4.NewReader(b)
	d, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(s, d) {
		t.Fatalf("decoded string mismatch\ngot: %q\nexp: %q", string(s), string(d))
	}
}
