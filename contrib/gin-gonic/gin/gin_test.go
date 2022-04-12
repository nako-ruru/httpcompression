package gin_test

import (
	ginhc "github.com/CAFxX/httpcompression/contrib/gin-gonic/gin"
	"github.com/gin-gonic/gin"

	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGin(t *testing.T) {
	r := gin.New()
	r.Use(ginhc.CompressResponse())
	r.GET("/", func(c *gin.Context) {
		c.Writer.WriteString(strings.Repeat("x", 1000))
	})

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip,br")
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("got code %d", res.Code)
	}
	if ae := res.Header().Get("Content-Encoding"); ae != "gzip" && ae != "br" {
		t.Fatalf("got content-encoding %q", ae)
	}
}
