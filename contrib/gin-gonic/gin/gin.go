package gin

import (
	"github.com/CAFxX/httpcompression"
	"github.com/gin-gonic/gin"
	wraphh "github.com/turtlemonvh/gin-wraphh"
)

func CompressResponse(opts ...httpcompression.Option) gin.HandlerFunc {
	hc, err := httpcompression.DefaultAdapter(opts...)
	if err != nil {
		panic(err)
	}
	return wraphh.WrapHH(hc)
}
