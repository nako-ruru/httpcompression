//go:build race
// +build race

package flate_test

import (
	"testing"

	"github.com/CAFxX/httpcompression/contrib/compress/flate"
	"github.com/CAFxX/httpcompression/contrib/internal"
)

func TestZstdRace(t *testing.T) {
	t.Parallel()
	c, _ := flate.New(flate.Options{})
	internal.RaceTestCompressionProvider(c, 100)
}
