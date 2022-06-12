// +build race

package brotli_test

import (
	"testing"

	"github.com/nako-ruru/httpcompression/contrib/andybalholm/brotli"
	"github.com/nako-ruru/httpcompression/contrib/internal"
)

func TestZstdRace(t *testing.T) {
	t.Parallel()
	c, _ := brotli.New(brotli.Options{})
	internal.RaceTestCompressionProvider(c, 100)
}
