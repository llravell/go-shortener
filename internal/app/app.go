package app

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/llravell/go-shortener/internal/rest"
	"github.com/llravell/go-shortener/internal/rest/middleware"
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
	urlUseCase    *usecase.URLUseCase
	healthUseCase *usecase.HealthUseCase
	router        chi.Router
	log           *zerolog.Logger
	addr          string
	jwtSecret     string
	isDebug       bool
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

func IsDebug(isDebug bool) Option {
	return func(app *App) {
		app.isDebug = isDebug
	}
}

func New(
	urlUseCase *usecase.URLUseCase,
	healthUseCase *usecase.HealthUseCase,
	log *zerolog.Logger,
	opts ...Option,
) *App {
	app := &App{
		urlUseCase:    urlUseCase,
		healthUseCase: healthUseCase,
		log:           log,
		router:        chi.NewRouter(),
	}

	for _, opt := range opts {
		opt(app)
	}

	return app
}

func (app *App) Run() {
	auth := middleware.NewAuth(app.jwtSecret, app.log)
	healthRoutes := rest.NewHealthRoutes(app.healthUseCase, app.log)
	urlRoutes := rest.NewURLRoutes(app.urlUseCase, auth, app.log)

	app.router.Use(middleware.LoggerMiddleware(app.log))
	healthRoutes.Apply(app.router)
	urlRoutes.Apply(app.router)

	if app.isDebug {
		app.router.Mount("/debug", chiMiddleware.Profiler())
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	serverNotify := make(chan error, 1)
	go func() {
		serverNotify <- startServer(app.addr, app.router)
		close(serverNotify)
	}()

	app.log.Info().
		Str("addr", app.addr).
		Msgf("starting shortener server on '%s'", app.addr)

	select {
	case s := <-interrupt:
		app.log.Info().Str("signal", s.String()).Msg("interrupt")
	case err := <-serverNotify:
		app.log.Error().Err(err).Msg("shortener server has been closed")
	}
}
