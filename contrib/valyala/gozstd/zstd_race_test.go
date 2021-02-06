// +build race

package zstd_test

import (
	"testing"

	"github.com/CAFxX/httpcompression/contrib/internal"
	zstd "github.com/CAFxX/httpcompression/contrib/valyala/gozstd"
)

func TestZstdRace(t *testing.T) {
	t.Parallel()
	c, _ := zstd.New(nil)
	internal.RaceTestCompressionProvider(c, 100)
}
