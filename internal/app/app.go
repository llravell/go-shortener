package app

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"

	"github.com/llravell/go-shortener/internal/rest"
	"github.com/llravell/go-shortener/internal/rest/middleware"
	"github.com/llravell/go-shortener/internal/usecase"
)

func startServer(addr string, handler http.Handler, https bool) error {
	server := http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	}

	if https {
		return server.ListenAndServeTLS("", "")
	}

	return server.ListenAndServe()
}

// Option дополнительная опция приложения.
type Option func(app *App)

// App приложение.
type App struct {
	urlUseCase    *usecase.URLUseCase
	healthUseCase *usecase.HealthUseCase
	statsUseCase  *usecase.StatsUseCase
	router        chi.Router
	log           *zerolog.Logger
	addr          string
	jwtSecret     string
	isDebug       bool
	httpsEnabled  bool
}

// Addr устанавливает адрес, на котором будет запускаться http сервер.
func Addr(addr string) Option {
	return func(app *App) {
		app.addr = addr
	}
}

// JWTSecret устанавливает строку, которая будет использоваться для генерации JWT.
func JWTSecret(secret string) Option {
	return func(app *App) {
		app.jwtSecret = secret
	}
}

// IsDebug устанавливает режим, в котором запущенно приложение.
func IsDebug(isDebug bool) Option {
	return func(app *App) {
		app.isDebug = isDebug
	}
}

// HTTPSEnabled запускает сервер по https.
func HTTPSEnabled(enabled bool) Option {
	return func(app *App) {
		app.httpsEnabled = enabled
	}
}

// New создает инстанс приложения.
func New(
	urlUseCase *usecase.URLUseCase,
	healthUseCase *usecase.HealthUseCase,
	statsUseCase *usecase.StatsUseCase,
	log *zerolog.Logger,
	opts ...Option,
) *App {
	app := &App{
		urlUseCase:    urlUseCase,
		healthUseCase: healthUseCase,
		statsUseCase:  statsUseCase,
		log:           log,
		router:        chi.NewRouter(),
	}

	for _, opt := range opts {
		opt(app)
	}

	return app
}

// Run инициализирует роуты и запускает http сервер.
func (app *App) Run() {
	auth := middleware.NewAuth(app.jwtSecret, app.log)
	healthRoutes := rest.NewHealthRoutes(app.healthUseCase, app.log)
	statsRoutes := rest.NewStatsRoutes(app.statsUseCase, app.log)
	urlRoutes := rest.NewURLRoutes(app.urlUseCase, auth, app.log)

	app.router.Use(middleware.LoggerMiddleware(app.log))
	healthRoutes.Apply(app.router)
	urlRoutes.Apply(app.router)

	app.router.Route("/api/internal", func(r chi.Router) {
		statsRoutes.Apply(r)
	})

	if app.isDebug {
		app.router.Mount("/debug", chiMiddleware.Profiler())
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	serverNotify := make(chan error, 1)
	go func() {
		serverNotify <- startServer(app.addr, app.router, app.httpsEnabled)
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
