// +build race

package cbrotli_test

import (
	"testing"

	"github.com/CAFxX/httpcompression/contrib/google/cbrotli"
	"github.com/CAFxX/httpcompression/contrib/internal"
	gcbrotli "github.com/google/brotli/go/cbrotli"
)

func TestZstdRace(t *testing.T) {
	t.Parallel()
	c, _ := cbrotli.New(gcbrotli.WriterOptions{})
	internal.RaceTestCompressionProvider(c, 100)
}
