//go:build race
// +build race

package flate_test

import (
	"testing"

	"github.com/CAFxX/httpcompression/contrib/internal"
	"github.com/CAFxX/httpcompression/contrib/klauspost/flate"
)

func TestDeflateRace(t *testing.T) {
	t.Parallel()
	c, _ := flate.New(flate.Options{})
	internal.RaceTestCompressionProvider(c, 100)
}
