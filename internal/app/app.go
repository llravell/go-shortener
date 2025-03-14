package app

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"

	"github.com/llravell/go-shortener/internal/grpc/interceptors"
	grpcServer "github.com/llravell/go-shortener/internal/grpc/server"
	pb "github.com/llravell/go-shortener/internal/proto"
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
	grpcAddr      string
	jwtSecret     string
	isDebug       bool
	httpsEnabled  bool
	trustedSubnet *net.IPNet
}

// Addr устанавливает адрес, на котором будет запускаться http сервер.
func Addr(addr string) Option {
	return func(app *App) {
		app.addr = addr
	}
}

// GRPCAddr устанавливает адрес, на котором будет запускаться grpc сервер.
func GRPCAddr(addr string) Option {
	return func(app *App) {
		app.grpcAddr = addr
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

// TrustedSubnet доверенная подсеть для доступа ко внутренним роутам.
func TrustedSubnet(subnet string) Option {
	return func(app *App) {
		if len(subnet) == 0 {
			return
		}

		_, trustedSubnet, err := net.ParseCIDR(subnet)
		if err != nil {
			return
		}

		app.trustedSubnet = trustedSubnet
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

// RunHTTP инициализирует роуты и запускает http сервер.
func (app *App) RunHTTP() {
	auth := middleware.NewAuth(app.jwtSecret, app.log)
	healthRoutes := rest.NewHealthRoutes(app.healthUseCase, app.log)
	urlRoutes := rest.NewURLRoutes(app.urlUseCase, auth, app.log)

	app.router.Use(middleware.LoggerMiddleware(app.log))
	healthRoutes.Apply(app.router)
	urlRoutes.Apply(app.router)

	if app.isDebug {
		app.router.Mount("/debug", chiMiddleware.Profiler())
	}

	if app.trustedSubnet != nil {
		statsRoutes := rest.NewStatsRoutes(app.statsUseCase, app.log)
		app.router.Route("/api/internal", func(r chi.Router) {
			r.Use(middleware.NetGuardMiddleware(app.trustedSubnet))

			statsRoutes.Apply(r)
		})
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
		Msgf("starting shortener http server on '%s'", app.addr)

	select {
	case s := <-interrupt:
		app.log.Info().Str("signal", s.String()).Msg("interrupt")
	case err := <-serverNotify:
		app.log.Error().Err(err).Msg("shortener http server has been closed")
	}
}

// RunGRPC запускает grpc сервер.
func (app *App) RunGRPC() {
	listen, err := net.Listen("tcp", app.grpcAddr)
	if err != nil {
		app.log.Error().Err(err).Msg("grpc server starting failed")

		return
	}

	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}
	loggingInterceptor := logging.UnaryServerInterceptor(interceptors.InterceptorLogger(app.log), loggingOpts...)

	shortenerServer := grpcServer.NewShortenerServer(app.urlUseCase, app.log)

	s := grpc.NewServer(grpc.UnaryInterceptor(loggingInterceptor))
	pb.RegisterShortenerServer(s, shortenerServer)

	app.log.Info().
		Str("grpcAddr", app.grpcAddr).
		Msgf("starting shortener grpc server on '%s'", app.grpcAddr)

	if err := s.Serve(listen); err != nil {
		app.log.Error().Err(err).Msg("shortener grpc server has been closed")
	}
}

// Run запускает приложение.
func (app *App) Run() {
	var wg sync.WaitGroup

	if len(app.addr) > 0 {
		wg.Add(1)

		go func() {
			app.RunHTTP()
			wg.Done()
		}()
	}

	if len(app.grpcAddr) > 0 {
		wg.Add(1)

		go func() {
			app.RunGRPC()
			wg.Done()
		}()
	}

	wg.Wait()
}
