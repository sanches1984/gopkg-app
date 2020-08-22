package app

import (
	"context"
	"encoding/base64"
	"fmt"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/opentracing/opentracing-go"
	"github.com/severgroup-tt/gopkg-app/middleware"
	"google.golang.org/grpc/reflection"
	"net"
	"net/http"
	"os"
	"regexp"
	"syscall"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/severgroup-tt/gopkg-app/client/sentry"
	"github.com/severgroup-tt/gopkg-app/closer"
	"github.com/severgroup-tt/gopkg-app/metrics"
	swaggerui "github.com/severgroup-tt/gopkg-app/swagger"
	pkgtransport "github.com/severgroup-tt/gopkg-app/transport"
	pkgvalidator "github.com/severgroup-tt/gopkg-app/validator"
	validatorerr "github.com/severgroup-tt/gopkg-app/validator/errors"
	validatormw "github.com/severgroup-tt/gopkg-app/validator/middleware"
	errors "github.com/severgroup-tt/gopkg-errors"
	errgrpc "github.com/severgroup-tt/gopkg-errors/grpc"
	errhttp "github.com/severgroup-tt/gopkg-errors/http"
	errmw "github.com/severgroup-tt/gopkg-errors/middleware"
	logger "github.com/severgroup-tt/gopkg-logger"

	"github.com/go-chi/chi"
	chiwm "github.com/go-chi/chi/middleware"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/utrack/clay/v2/transport"
	"github.com/utrack/clay/v2/transport/swagger"
	"google.golang.org/grpc"
)

var (
	gracefulDelay   = time.Duration(3) * time.Second
	gracefulTimeout = time.Duration(10) * time.Second
)

type App struct {
	config Config

	httpServer        *chi.Mux
	httpListener      net.Listener
	httpAdminServer   *chi.Mux
	httpAdminListener net.Listener
	grpcServer        *grpc.Server
	grpcListener      net.Listener

	unaryInterceptor []grpc.UnaryServerInterceptor
	publicMiddleware []func(http.Handler) http.Handler

	tracer *opentracing.Tracer

	publicCloser *closer.Closer

	favicon             []byte
	adminURLPrefix      string
	customPublicHandler []PublicHandler
	customPublicCloser  PublicCloserFnMap
	customSwaggerOption []swagger.Option
	customEnablePprof   bool
}

func NewApp(ctx context.Context, config Config, option ...OptionFn) (*App, error) {
	pkgtransport.Override(nil)
	metrics.AddBasicCollector(config.Name)

	favicon, _ := base64.StdEncoding.DecodeString("AAABAAEAEBAAAAEAIABoBAAAFgAAACgAAAAQAAAAIAAAAAEAIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA//////////////////////////////////7+//3+/v/9/v7//v79//7++v/+/v3//f7+//3+/v8AAAAA//////////////////////////////////////7+/v/8/v3/+vfu//PPlv/vumz/8cF4//bmxv/9/v3//f79///////////////////////////////////////+/v7//P77//TGg//2s1X/+rRS//q0Uv/ws1v/+OvR//7+/f///////////////////////////////////////v7+//n05P/0slj/+rNT//mzU//6s1L/+rNT//PPmP/+/v7///////////////////////////////////////7+/v/48+P/87NX//i0Uv/5tFP/+LRS//mzVP/zz5X//v79///////////////////////////////////////+/v7/+/37//LDgf/4s1P/+LRT//m0Uv/3s1n/9erQ//3+/v///////////////////////////////////////P7+//3+/f/59+r/8syS//S4Z//zv3P/9ePF//z+/P/+/v3///////////////////////////////////////7+/v/+/v7//f79//z+/P/+/fn//f75//3+/f/+/v3//v7+//3+/v/8/f3/9vv7/7zq9/+d4Pf/uen3//T7/P/7/f3//f7+//v8/f/wz9j/6Ka0/+q0wP/68PP/+/7+//7+/v/7/v7/+f38/4rX9P9VyPX/Usj3/1HJ9f+F1fX/9fz9//3+/v/or7r/42B3/+dfdv/oX3f/3nSH//ns8P/9/v7//f3+/8zv+f9UyPX/Tsn3/1HI9/9QyPj/Usj1/8ft+P/58PP/32V6/+lfd//nX3f/5193/+dfd//nqrb//v7+//v9/f+66vj/Usj3/1HI9/9RyPf/Ucj3/1LI9/+x5vn/8uLm/+Bgd//nX3f/5193/+dfd//oX3f/5Zem//7+/v/9/v7/3vT6/1jK8/9Ryff/Ucn3/1LI9/9TyvT/2PL7//z4+v/gb4T/5l92/+dfd//nX3f/4193/+66xf/+/v7//v7+//z9/v+z5vj/XMvz/1XI9v9Zy/L/quL3//r9/v/9/v7/79LY/+Btgf/lX3b/4mF3/+OWpP/7+fv//v7+///////+/v7//P7+/+b4+//N8Pn/5fb7//v9/P/9/v7/+v7+//z8/f/68/b/79Xb//Ti5//8/f3//f7+//3+/v8AAAAA//////7+/v/7/v7//P7+//3+/v/9/v7/+v7+//z+/v/7/v7//P7+//3+///9/v7//P7+//z+/v8AAAAAgAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgAEAAA==")
	a := &App{
		config:             config,
		favicon:            favicon,
		unaryInterceptor:   getDefaultUnaryInterceptor(config.Name),
		publicMiddleware:   getDefaultPublicMiddleware(config.Version),
		publicCloser:       closer.New(syscall.SIGTERM, syscall.SIGINT),
		customPublicCloser: make(PublicCloserFnMap),
	}

	if err := a.initServers(); err != nil {
		return nil, err
	}

	for _, optFn := range option {
		if err := optFn(a); err != nil {
			return nil, err
		}
	}

	return a, nil
}

func (a *App) Run(impl ...transport.Service) {
	var descs []transport.ServiceDesc
	for _, i := range impl {
		descs = append(descs, i.GetDescription())
	}
	implDesc := transport.NewCompoundServiceDesc(descs...)
	implDesc.Apply(transport.WithUnaryInterceptor(grpc_middleware.ChainUnaryServer(a.unaryInterceptor...)))
	a.runServers(implDesc)
}

func GracefulDelay(serviceName string) {
	logger.Info(logger.App, serviceName+": waiting stop of traffic")
	time.Sleep(gracefulDelay)
	logger.Info(logger.App, serviceName+": shutting down")
}

func (a *App) runServers(impl *transport.CompoundServiceDesc) {
	if a.grpcListener != nil {
		a.grpcServer = grpc.NewServer(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(a.unaryInterceptor...)))
		impl.RegisterGRPC(a.grpcServer)
		reflection.Register(a.grpcServer)
		a.runGRPC()
	}

	if a.httpListener != nil {
		a.httpServer.Use(a.publicMiddleware...)
		impl.RegisterHTTP(a.httpServer)
		a.httpServer.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "text/html")
			text := a.config.Name + " " + a.config.Version
			_, _ = w.Write([]byte("<html><head><title>" + text + `</title></head><body><h1 style="text-align: center; margin-top: 200px; font-size: 50px;">` + text + "</h1></body></html>"))
		})
		a.httpServer.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "image/x-icon")
			_, _ = w.Write(a.favicon)
		})
		if a.customEnablePprof {
			logger.Info(logger.App, "PPROF enabled")
			a.httpServer.Mount("/debug", chiwm.Profiler())
		}
		for _, h := range a.customPublicHandler {
			a.httpServer.MethodFunc(h.Method, h.Pattern, h.NewHandlerFuncWithMiddleware())
		}
		a.runPublicHTTP()
	}

	if a.httpAdminListener != nil {
		a.initAdminHandlers(impl)
		a.runAdminHTTP()
	}

	for name, fn := range a.customPublicCloser {
		a.publicCloser.Add(name, fn)
	}

	// Wait signal and close all resources
	a.publicCloser.Wait()
	// Close all other resources from globalCloser
	closer.CloseAll()
}

func getDefaultUnaryInterceptor(appName string) []grpc.UnaryServerInterceptor {
	errConverters := []errors.ErrorConverter{
		validatorerr.Converter(),
		errhttp.Converter(appName),
		errgrpc.Converter(appName),
	}
	return []grpc.UnaryServerInterceptor{
		grpc_ctxtags.UnaryServerInterceptor(),
		grpc_prometheus.UnaryServerInterceptor,
		errmw.NewConvertErrorsServerInterceptor(errConverters, &metrics.CountError),
		validatormw.NewValidateServerInterceptor(pkgvalidator.New()),
		middleware.NewLogInterceptor(),
		grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(func(data interface{}) (err error) {
			sentry.Panic(data)
			return nil
		})),
	}
}

func getDefaultPublicMiddleware(appVersion string) []func(http.Handler) http.Handler {
	ret := make([]func(http.Handler) http.Handler, 0, 10)
	ret = append(ret, middleware.NewTimingMiddleware()...)
	ret = append(ret,
		middleware.NewHeartbeatMiddleware(),
		middleware.NewCorsMiddleware(),
		middleware.NewRequestIdMiddleware(),
		middleware.NewLogMiddleware(),
		middleware.NewNoCacheMiddleware(),
		middleware.NewVersionMiddleware(appVersion),
	)
	return ret
}

func (a *App) initServers() error {
	logger.Info(logger.App, "App '%s' version '%s' in %s started",
		a.config.Name,
		a.config.Version,
		a.config.Env,
	)

	if a.config.Listener.HttpPort != 0 {
		logger.Info(logger.App, "Starting public HTTP listener at %s:%d", a.config.Listener.Host, a.config.Listener.HttpPort)
		httpListener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", a.config.Listener.Host, a.config.Listener.HttpPort))
		if err != nil {
			return err
		}
		a.httpListener = httpListener
		a.httpServer = chi.NewMux()
	}

	if a.config.Listener.HttpAdminPort != 0 {
		logger.Info(logger.App, "Starting admin HTTP listener at %s:%d", a.config.Listener.Host, a.config.Listener.HttpAdminPort)
		httpAdminListener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", a.config.Listener.Host, a.config.Listener.HttpAdminPort))
		if err != nil {
			return err
		}
		a.httpAdminListener = httpAdminListener
		a.httpAdminServer = chi.NewMux()
	}

	if a.config.Listener.GrpcPort != 0 {
		logger.Info(logger.App, "Starting GRPC listener at %s:%d", a.config.Listener.Host, a.config.Listener.GrpcPort)
		grpcListener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", a.config.Listener.Host, a.config.Listener.GrpcPort))
		if err != nil {
			return err
		}
		a.grpcListener = grpcListener
	}

	return nil
}

func (a *App) initAdminHandlers(implDesc *transport.CompoundServiceDesc) {
	urlPrefix := a.adminURLPrefix

	// table of contents
	a.httpAdminServer.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, urlPrefix+"/", 301)
	})
	a.httpAdminServer.Get("/docs/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, urlPrefix+"/", 301)
	})
	a.httpAdminServer.Get(urlPrefix+"/", func(w http.ResponseWriter, r *http.Request) {
		body := "<h1>Table of contents</h1><ul>"
		if a.config.Listener.HttpPort != 0 {
			body += `<li><a href="` + urlPrefix + `/docs/rest/">REST documentation</a></li>`
		}
		if a.config.Listener.GrpcPort != 0 {
			body += `<li><a href="` + urlPrefix + `/docs/grpc/">GRPC documentation</a></li>`
		}
		body += `<li><a href="` + urlPrefix + `/metrics">Metrics</a></li>`
		body += `</ul>`
		_, _ = w.Write([]byte(body))
	})

	// metrics
	a.httpAdminServer.Mount("/metrics", metrics.Metrics())

	// grpc documentation
	if a.config.Listener.GrpcPort != 0 {
		a.httpAdminServer.Get("/docs/grpc", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, urlPrefix+"/docs/grpc/", 301)
		})
		a.httpAdminServer.HandleFunc("/docs/grpc/", func(w http.ResponseWriter, r *http.Request) {
			filePath := "docs/grpc/index.html"
			_, err := os.Stat(filePath)
			if os.IsNotExist(err) {
				_, _ = w.Write([]byte(`<div>Insert into Makefile and execute:</div>
<pre>.PHONY: bin-deps
bin-deps: ; $(info $(M) install bin depends…) @ ## Install bin depends
	...
	GOBIN=$(LOCAL_BIN) $(GO_EXEC) install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc

.PHONY: doc-grpc
doc-grpc: bin-deps ; $(info $(M) generate grpc docs…) @ ## Generate GRPC documentation
	protoc \
		--plugin=protoc-gen-doc=$(LOCAL_BIN)/protoc-gen-doc \
		-I./api/:./vendor.pb \
		--doc_out=./docs/grpc \
		--doc_opt=html,index.html ./api/*.proto</pre>`))
				return
			}
			http.ServeFile(w, r, filePath)
		})
	}

	// swagger
	if a.config.Listener.HttpPort != 0 {
		a.httpAdminServer.Get("/docs/rest", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, urlPrefix+"/docs/rest/", 301)
		})
		a.httpAdminServer.Mount("/docs/rest/", http.StripPrefix("/docs/rest", swaggerui.NewHTTPHandler()))
		a.httpAdminServer.Get("/docs/rest/swagger.json", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, urlPrefix+"/swagger.json", 301)
		})
		removeSchemeRE := regexp.MustCompile("^https?://")
		hostWithoutScheme := removeSchemeRE.ReplaceAllString(a.config.Host, "")
		a.httpAdminServer.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-MimeType", "application/json")
			o := []swagger.Option{
				swagger.WithHost(hostWithoutScheme),
				swagger.WithTitle(a.config.Name),
				swagger.WithVersion(a.config.Version),
				pkgtransport.SetIntegerTypeForInt64(),
				pkgtransport.SetDeprecatedFromSummary(),
				pkgtransport.SetErrorResponse(),
				pkgtransport.SetNameSnakeCase(),
			}
			o = append(o, a.customSwaggerOption...)
			_, _ = w.Write(implDesc.SwaggerDef(o...))
		})
	}
}

func (a *App) runGRPC() {
	go func() {
		if err := a.grpcServer.Serve(a.grpcListener); err != nil {
			logger.Error(logger.App, "grpc: %s", err)
			a.publicCloser.CloseAll()
		}
	}()
	a.publicCloser.Add("grpc", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
		defer cancel()

		GracefulDelay("grpc")

		done := make(chan struct{})
		go func() {
			a.grpcServer.GracefulStop()
			close(done)
		}()
		select {
		case <-done:
			logger.Info(logger.App, "grpc: gracefully stopped")
		case <-ctx.Done():
			err := errors.Internal.Err(context.Background(), "grpc: error during shutdown server").
				WithLogKV("error", ctx.Err())
			a.grpcServer.Stop()
			return errors.Internal.Err(context.Background(), "grpc: force stopped").
				WithLogKV("error", err)
		}
		return nil
	})
}

func (a *App) runPublicHTTP() {
	publicServer := &http.Server{Handler: a.httpServer}
	go func() {
		if err := publicServer.Serve(a.httpListener); err != nil {
			logger.Info(logger.App, "http.public: %s", err)
			a.publicCloser.CloseAll()
		}
	}()
	a.publicCloser.Add("http.public", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
		defer cancel()

		GracefulDelay("http.public")

		publicServer.SetKeepAlivesEnabled(false)
		if err := publicServer.Shutdown(ctx); err != nil {
			return errors.Internal.Err(context.Background(), "http.public: error during shutdown").
				WithLogKV("error", err)
		}
		logger.Info(logger.App, "http.public: gracefully stopped")
		return nil
	})
}

func (a *App) runAdminHTTP() {
	adminServer := &http.Server{Handler: a.httpAdminServer}
	go func() {
		if err := adminServer.Serve(a.httpAdminListener); err != nil {
			logger.Info(logger.App, "admin.public: %s", err)
			a.publicCloser.CloseAll()
		}
	}()
	a.publicCloser.Add("admin.public", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
		defer cancel()

		GracefulDelay("admin.public")

		adminServer.SetKeepAlivesEnabled(false)
		if err := adminServer.Shutdown(ctx); err != nil {
			return errors.Internal.Err(context.Background(), "admin.public: error during shutdown").
				WithLogKV("error", err)
		}
		logger.Info(logger.App, "admin.public: gracefully stopped")
		return nil
	})
}
