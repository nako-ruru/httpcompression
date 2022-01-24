package zlib_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	stdzlib "compress/zlib"

	"github.com/CAFxX/httpcompression"
	"github.com/CAFxX/httpcompression/contrib/compress/zlib"
)

var _ httpcompression.CompressorProvider = &zlib.Compressor{}

func TestDeflate(t *testing.T) {
	t.Parallel()

	s := []byte("hello world!")

	c, err := zlib.New(zlib.Options{Dictionary: s})
	if err != nil {
		t.Fatal(err)
	}
	b := &bytes.Buffer{}
	w := c.Get(b)
	w.Write(s)
	w.Close()

	r, err := stdzlib.NewReaderDict(b, s)
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
