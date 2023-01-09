package web

import (
	"net/http"

	"github.com/avalonbits/shortener/service"
	"github.com/labstack/echo/v4"
)

type Handlers struct {
	svc *service.Shortener
}

func New(svc *service.Shortener) *Handlers {
	return &Handlers{
		svc: svc,
	}
}

func (h *Handlers) Root(c echo.Context) error {
	return c.Render(http.StatusOK, "root", nil)
}
