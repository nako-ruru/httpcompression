//go:build race
// +build race

package zlib_test

import (
	"testing"

	"github.com/nako-ruru/httpcompression/contrib/compress/zlib"
	"github.com/nako-ruru/httpcompression/contrib/internal"
)

func TestZstdRace(t *testing.T) {
	t.Parallel()
	c, _ := zlib.New(zlib.Options{})
	internal.RaceTestCompressionProvider(c, 100)
}
