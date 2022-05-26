package fiber

import (
	"github.com/CAFxX/httpcompression"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
)

func Adapter(opts ...httpcompression.Option) (fiber.Handler, error) {
	mw, err := httpcompression.Adapter(opts...)
	if err != nil {
		return nil, err
	}
	return adaptor.HTTPMiddleware(mw), nil
}

func DefaultAdapter(opts ...httpcompression.Option) (fiber.Handler, error) {
	mw, err := httpcompression.DefaultAdapter(opts...)
	if err != nil {
		return nil, err
	}
	return adaptor.HTTPMiddleware(mw), nil
}
