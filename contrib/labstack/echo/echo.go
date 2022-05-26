package echo

import (
	"github.com/CAFxX/httpcompression"
	"github.com/labstack/echo"
)

func Adapter(opts ...httpcompression.Option) (echo.MiddlewareFunc, error) {
	mw, err := httpcompression.Adapter(opts...)
	if err != nil {
		return nil, err
	}
	return echo.WrapMiddleware(mw), nil
}

func DefaultAdapter(opts ...httpcompression.Option) (echo.MiddlewareFunc, error) {
	mw, err := httpcompression.DefaultAdapter(opts...)
	if err != nil {
		return nil, err
	}
	return echo.WrapMiddleware(mw), nil
}
