package app

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/llravell/go-shortener/internal/rest"
	"github.com/llravell/go-shortener/internal/usecase"
	"github.com/rs/zerolog"
)

func startServer(addr string, handler http.Handler) error {
	server := http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	}

	return server.ListenAndServe()
}

type Option func(app *App)

type App struct {
	urlUseCase       *usecase.URLUseCase
	urlDeleteUseCase *usecase.URLDeleteUseCase
	healthUseCase    *usecase.HealthUseCase
	log              zerolog.Logger
	addr             string
	jwtSecret        string
}

func Addr(addr string) Option {
	return func(app *App) {
		app.addr = addr
	}
}

func JWTSecret(secret string) Option {
	return func(app *App) {
		app.jwtSecret = secret
	}
}

func New(
	urlUseCase *usecase.URLUseCase,
	urlDeleteUseCase *usecase.URLDeleteUseCase,
	healthUseCase *usecase.HealthUseCase,
	log zerolog.Logger,
	opts ...Option,
) *App {
	app := &App{
		urlUseCase:       urlUseCase,
		urlDeleteUseCase: urlDeleteUseCase,
		healthUseCase:    healthUseCase,
		log:              log,
	}

	for _, opt := range opts {
		opt(app)
	}

	return app
}

func (app *App) Run() {
	router := rest.NewRouter(
		app.urlUseCase,
		app.urlDeleteUseCase,
		app.healthUseCase,
		app.jwtSecret,
		app.log,
	)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	serverNorify := make(chan error, 1)
	go func() {
		serverNorify <- startServer(app.addr, router)
		close(serverNorify)
	}()

	app.log.Info().
		Str("addr", app.addr).
		Msgf("starting shortener server on '%s'", app.addr)

	go app.urlDeleteUseCase.ProcessQueue()
	defer app.urlDeleteUseCase.Close()

	select {
	case s := <-interrupt:
		app.log.Info().Str("signal", s.String()).Msg("interrupt")
	case err := <-serverNorify:
		app.log.Error().Err(err).Msg("shortener server has been closed")
	}
}
