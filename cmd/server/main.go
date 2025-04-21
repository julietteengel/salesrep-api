package main

import (
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.uber.org/fx"

	// Autoload .env file
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	// Both formats are needed or subseconds will be lost. https://github.com/rs/zerolog/issues/114
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05.999999"})
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05.999999"

	fx.New(
		fx.Provide(
			DatabaseProvider,
			NewHTTPServer,
			NewHTTPRouter,
		),
		fx.Invoke(
			func(_ *http.Server) {},
		),
		fx.StopTimeout(1*time.Minute),
	).Run()
}
