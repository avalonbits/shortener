package web

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/avalonbits/shortener/service"
	"github.com/labstack/echo/v4"
)

type domain string

func (d domain) IsDev() bool {
	return strings.HasPrefix(string(d), "localhost")
}

func (d domain) URL(path ...string) string {
	var result string
	var err error
	if d.IsDev() {
		result, err = url.JoinPath("http://"+string(d), path...)
	} else {
		result, err = url.JoinPath("https://"+string(d), path...)
	}

	if err != nil {
		// If this happens, it was a programming error because we expect that path to be
		// programmer provided, not user provided.
		panic(err)
	}
	return result
}

type Handlers struct {
	svc *service.Shortener
	dom domain
}

func New(svc *service.Shortener, urlDomain string) *Handlers {
	return &Handlers{
		svc: svc,
		dom: domain(urlDomain),
	}
}

func (h *Handlers) Root(c echo.Context) error {
	return c.Render(http.StatusOK, "root", nil)
}

type urlParam struct {
	FormURL string `form:"url"`
}

type urlResponse struct {
	URL string `json:"url"`
}

func (h *Handlers) CreateShortURL(c echo.Context) error {
	var input urlParam
	if err := c.Bind(&input); err != nil {
		return c.String(http.StatusBadRequest, "invalid url")
	}

	if input.FormURL == "" {
		return c.String(http.StatusBadRequest, "missing an url")
	}

	short, err := h.svc.ShortNameFor(c.Request().Context(), input.FormURL)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	shortURL := h.dom.URL(short)
	return c.Render(http.StatusOK, "short_url", map[string]string{"ShortURL": shortURL})
}
