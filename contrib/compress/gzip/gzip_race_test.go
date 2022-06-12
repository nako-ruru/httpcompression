//go:build race
// +build race

package gzip_test

import (
	"testing"

	"github.com/nako-ruru/httpcompression/contrib/compress/gzip"
	"github.com/nako-ruru/httpcompression/contrib/internal"
)

func TestZstdRace(t *testing.T) {
	t.Parallel()
	c, _ := gzip.New(gzip.Options{})
	internal.RaceTestCompressionProvider(c, 100)
}
