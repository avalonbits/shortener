package main

import (
	"context"
	"crypto/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/avalonbits/shortener/embed"
	"github.com/avalonbits/shortener/endoints/web"
	"github.com/avalonbits/shortener/service"
	"github.com/avalonbits/shortener/storage"
	"github.com/labstack/echo/v4"
)

func main() {
	e, bind, port := setup()
	runServer(e, bind, port)
}

func runServer(e *echo.Echo, bind, port string) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		if err := e.Start(bind + ":" + port); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal(err)
		}
	}()
	<-sigc

	if err := e.Shutdown(context.Background()); err != nil {
		e.Logger.Fatal(err)
	}
}

func setup() (*echo.Echo, string, string) {
	db := storage.ProdDB("/tmp/testshort.db")
	queries := storage.New(db)
	e := echo.New()

	templates := embed.Templates()
	e.Renderer = templates
	handlers := web.New(
		service.NewShortener(queries, rand.Reader, 2 /*existsRetry=*/),
		"localhost:9001",
	)
	templates.NewView("root", "root.tmpl")
	e.GET("/", handlers.Root)
	e.POST("/short", handlers.CreateShortURL)

	return e, "localhost", "9001"
}
