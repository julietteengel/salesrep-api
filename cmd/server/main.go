package main

import (
	"context"
	"github.com/julietteengel/salesrep-api/internal/fx/controllers"
	"github.com/labstack/echo/v4"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"go.uber.org/fx"

	// Autoload .env file
	_ "github.com/joho/godotenv/autoload"
)

// @title Salesrep backend API
// @description This is the backend for Salesrep app
// @contact.name Juliette Engel
// @contact.email juliette.engel@skema.edu
func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05.999999"})
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05.999999"

	viper.AutomaticEnv()

	fx.New(
		// HTTP server
		fx.Provide(
			fx.Annotate(NewEchoServer, fx.ParamTags(`group:"controllers"`)),
		),

		// Database
		fx.Provide(
			NewGormDB,
		),

		controllers.PrivateControllers,

		fx.Invoke(StartEchoServer),
		fx.StopTimeout(1*time.Minute),
	).Run()
}

func StartEchoServer(lc fx.Lifecycle, e *echo.Echo) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := e.Start(":4000"); err != nil && err != http.ErrServerClosed {
					log.Fatal().Err(err).Msg("Echo server failed")
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info().Msg("Shutting down Echo server")
			return e.Shutdown(ctx)
		},
	})
}
