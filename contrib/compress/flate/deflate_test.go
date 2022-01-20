package flate_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	stdflate "compress/flate"

	"github.com/CAFxX/httpcompression"
	"github.com/CAFxX/httpcompression/contrib/compress/flate"
)

var _ httpcompression.CompressorProvider = &flate.Compressor{}

func TestDeflate(t *testing.T) {
	t.Parallel()

	s := []byte("hello world!")

	c, err := flate.New(flate.Options{Dictionary: s})
	if err != nil {
		t.Fatal(err)
	}
	b := &bytes.Buffer{}
	w := c.Get(b)
	w.Write(s)
	w.Close()

	r := stdflate.NewReaderDict(b, s)
	d, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(s, d) {
		t.Fatalf("decoded string mismatch\ngot: %q\nexp: %q", string(s), string(d))
	}
}
