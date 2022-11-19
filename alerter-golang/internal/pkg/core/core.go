package core

import (
	"fmt"
	"github.com/fatih/color"
	"golang.org/x/net/trace"
	"html/template"
	"net/http"
	"net/url"
	"runtime/debug"
	"time"

	"github.com/Project-Uranus-plus/all-eyes/configs"
	"github.com/Project-Uranus-plus/all-eyes/internal/code"
	"github.com/Project-Uranus-plus/all-eyes/pkg/env"
	"github.com/Project-Uranus-plus/all-eyes/pkg/errors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	cors "github.com/rs/cors/wrapper/gin"

	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

const _UI = `
`

type Option func(*option)

type option struct {
	disablePProf      bool
	disableSwagger    bool
	disablePrometheus bool
	enableCors        bool
	enableRate        bool
	enableOpenBrowser string
}

// WithDisablePProf 禁用 pprof
func WithDisablePProf() Option {
	return func(opt *option) {
		opt.disablePProf = true
	}
}

// WithDisableSwagger 禁用 swagger
func WithDisableSwagger() Option {
	return func(opt *option) {
		opt.disableSwagger = true
	}
}

// WithDisablePrometheus 禁用prometheus
func WithDisablePrometheus() Option {
	return func(opt *option) {
		opt.disablePrometheus = true
	}
}


// WithEnableOpenBrowser 启动后在浏览器中打开 uri
func WithEnableOpenBrowser(uri string) Option {
	return func(opt *option) {
		opt.enableOpenBrowser = uri
	}
}

// WithEnableCors 设置支持跨域
func WithEnableCors() Option {
	return func(opt *option) {
		opt.enableCors = true
	}
}

// WithEnableRate 设置支持限流
func WithEnableRate() Option {
	return func(opt *option) {
		opt.enableRate = true
	}
}

// DisableTraceLog 禁止记录日志
func DisableTraceLog(ctx Context) {
	ctx.disableTrace()
}

// DisableRecordMetrics 禁止记录指标
func DisableRecordMetrics(ctx Context) {
	ctx.disableRecordMetrics()
}

// AliasForRecordMetrics 对请求路径起个别名，用于记录指标。
// 如：Get /user/:username 这样的路径，因为 username 会有非常多的情况，这样记录指标非常不友好。
func AliasForRecordMetrics(path string) HandlerFunc {
	return func(ctx Context) {
		ctx.setAlias(path)
	}
}


// RouterGroup 包装gin的RouterGroup
type RouterGroup interface {
	Group(string, ...HandlerFunc) RouterGroup
	IRoutes
}

var _ IRoutes = (*router)(nil)

// IRoutes 包装gin的IRoutes
type IRoutes interface {
	Any(string, ...HandlerFunc)
	GET(string, ...HandlerFunc)
	POST(string, ...HandlerFunc)
	DELETE(string, ...HandlerFunc)
	PATCH(string, ...HandlerFunc)
	PUT(string, ...HandlerFunc)
	OPTIONS(string, ...HandlerFunc)
	HEAD(string, ...HandlerFunc)
}

type router struct {
	group *gin.RouterGroup
}

func (r *router) Group(relativePath string, handlers ...HandlerFunc) RouterGroup {
	group := r.group.Group(relativePath, wrapHandlers(handlers...)...)
	return &router{group: group}
}

func (r *router) Any(relativePath string, handlers ...HandlerFunc) {
	r.group.Any(relativePath, wrapHandlers(handlers...)...)
}

func (r *router) GET(relativePath string, handlers ...HandlerFunc) {
	r.group.GET(relativePath, wrapHandlers(handlers...)...)
}

func (r *router) POST(relativePath string, handlers ...HandlerFunc) {
	r.group.POST(relativePath, wrapHandlers(handlers...)...)
}

func (r *router) DELETE(relativePath string, handlers ...HandlerFunc) {
	r.group.DELETE(relativePath, wrapHandlers(handlers...)...)
}

func (r *router) PATCH(relativePath string, handlers ...HandlerFunc) {
	r.group.PATCH(relativePath, wrapHandlers(handlers...)...)
}

func (r *router) PUT(relativePath string, handlers ...HandlerFunc) {
	r.group.PUT(relativePath, wrapHandlers(handlers...)...)
}

func (r *router) OPTIONS(relativePath string, handlers ...HandlerFunc) {
	r.group.OPTIONS(relativePath, wrapHandlers(handlers...)...)
}

func (r *router) HEAD(relativePath string, handlers ...HandlerFunc) {
	r.group.HEAD(relativePath, wrapHandlers(handlers...)...)
}

func wrapHandlers(handlers ...HandlerFunc) []gin.HandlerFunc {
	funcs := make([]gin.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		handler := handler
		funcs[i] = func(c *gin.Context) {
			ctx := newContext(c)
			defer releaseContext(ctx)

			handler(ctx)
		}
	}

	return funcs
}

var _ Mux = (*mux)(nil)

// Mux http mux
type Mux interface {
	http.Handler
	Group(relativePath string, handlers ...HandlerFunc) RouterGroup
}

type mux struct {
	engine *gin.Engine
}

func (m *mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	m.engine.ServeHTTP(w, req)
}

func (m *mux) Group(relativePath string, handlers ...HandlerFunc) RouterGroup {
	return &router{
		group: m.engine.Group(relativePath, wrapHandlers(handlers...)...),
	}
}

func New(logger *zap.Logger, options ...Option) (Mux, error) {
	if logger == nil {
		return nil, errors.New("logger required")
	}

	gin.SetMode(gin.ReleaseMode)
	mux := &mux{
		engine: gin.New(),
	}

	fmt.Println(color.Blue(_UI))

	mux.engine.StaticFS("assets", http.FS(assets.Bootstrap))
	mux.engine.SetHTMLTemplate(template.Must(template.New("").ParseFS(assets.Templates, "templates/**/*")))

	// withoutTracePaths 这些请求，默认不记录日志
	withoutTracePaths := map[string]bool{
		"/metrics": true,

		"/debug/pprof/":             true,
		"/debug/pprof/cmdline":      true,
		"/debug/pprof/profile":      true,
		"/debug/pprof/symbol":       true,
		"/debug/pprof/trace":        true,
		"/debug/pprof/allocs":       true,
		"/debug/pprof/block":        true,
		"/debug/pprof/goroutine":    true,
		"/debug/pprof/heap":         true,
		"/debug/pprof/mutex":        true,
		"/debug/pprof/threadcreate": true,

		"/favicon.ico": true,

		"/system/health": true,
	}

	opt := new(option)
	for _, f := range options {
		f(opt)
	}

	if !opt.disablePProf {
		if !env.Active().IsPro() {
			pprof.Register(mux.engine) // register pprof to gin
		}
	}

	if !opt.disableSwagger {
		if !env.Active().IsPro() {
			mux.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler)) // register swagger
		}
	}

	if !opt.disablePrometheus {
		mux.engine.GET("/metrics", gin.WrapH(promhttp.Handler())) // register prometheus
	}

	if opt.enableCors {
		mux.engine.Use(cors.New(cors.Options{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{
				http.MethodHead,
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
			},
			AllowedHeaders:     []string{"*"},
			AllowCredentials:   true,
			OptionsPassthrough: true,
		}))
	}


	// recover两次，防止处理时发生panic，尤其是在OnPanicNotify中。
	mux.engine.Use(func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("got panic", zap.String("panic", fmt.Sprintf("%+v", err)), zap.String("stack", string(debug.Stack())))
			}
		}()

		ctx.Next()
	})

	mux.engine.Use(func(ctx *gin.Context) {

		if ctx.Writer.Status() == http.StatusNotFound {
			return
		}

		ts := time.Now()

		context := newContext(ctx)
		defer releaseContext(context)

		context.init()
		context.setLogger(logger)
		context.ableRecordMetrics()

		if !withoutTracePaths[ctx.Request.URL.Path] {
			if traceId := context.GetHeader(trace.Header); traceId != "" {
				context.setTrace(trace.New(traceId))
			} else {
				context.setTrace(trace.New(""))
			}
		}

		defer func() {
			var (
				response        interface{}
				businessCode    int
				businessCodeMsg string
				abortErr        error
				traceId         string
				graphResponse   interface{}
			)

			if ct := context.Trace(); ct != nil {
				context.SetHeader(trace.Header, ct.ID())
				traceId = ct.ID()
			}

			// region 发生 Panic 异常发送告警提醒
			if err := recover(); err != nil {
				stackInfo := string(debug.Stack())
				logger.Error("got panic", zap.String("panic", fmt.Sprintf("%+v", err)), zap.String("stack", stackInfo))
				context.AbortWithError(Error(
					http.StatusInternalServerError,
					code.ServerError,
					code.Text(code.ServerError)),
				)

			}
			// endregion

			// region 发生错误，进行返回
			if ctx.IsAborted() {
				for i := range ctx.Errors {
					multierr.AppendInto(&abortErr, ctx.Errors[i])
				}

				if err := context.abortError(); err != nil { // customer err

					multierr.AppendInto(&abortErr, err.StackError())
					businessCode = err.BusinessCode()
					businessCodeMsg = err.Message()
					response = &code.Failure{
						Code:    businessCode,
						Message: businessCodeMsg,
					}
					ctx.JSON(err.HTTPCode(), response)
				}
			}
			// endregion

			// region 正确返回
			response = context.getPayload()
			if response != nil {
				ctx.JSON(http.StatusOK, response)
			}
			// endregion


			// endregion

			// region 记录日志
			var t *trace.Trace
			if x := context.Trace(); x != nil {
				t = x.(*trace.Trace)
			} else {
				return
			}

			decodedURL, _ := url.QueryUnescape(ctx.Request.URL.RequestURI())



			logger.Info("trace-log",
				zap.Any("method", ctx.Request.Method),
				zap.Any("path", decodedURL),
				zap.Any("http_code", ctx.Writer.Status()),
				zap.Any("business_code", businessCode),
				zap.Any("trace_info", t),
				zap.Error(abortErr),
			)
			// endregion
		}()

		ctx.Next()
	})

	if opt.enableRate {
		limiter := rate.NewLimiter(rate.Every(time.Second*1), configs.MaxRequestsPerSecond)
		mux.engine.Use(func(ctx *gin.Context) {
			context := newContext(ctx)
			defer releaseContext(context)

			if !limiter.Allow() {
				context.AbortWithError(Error(
					http.StatusTooManyRequests,
					code.TooManyRequests,
					code.Text(code.TooManyRequests)),
				)
				return
			}

			ctx.Next()
		})
	}

	mux.engine.NoMethod(wrapHandlers(DisableTraceLog)...)
	mux.engine.NoRoute(wrapHandlers(DisableTraceLog)...)

	system := mux.Group("/system")
	{
		// 健康检查
		system.GET("/health", func(ctx Context) {
			resp := &struct {
				Timestamp   time.Time `json:"timestamp"`
				Environment string    `json:"environment"`
				Host        string    `json:"host"`
				Status      string    `json:"status"`
			}{
				Timestamp:   time.Now(),
				Environment: env.Active().Value(),
				Host:        ctx.Host(),
				Status:      "ok",
			}
			ctx.Payload(resp)
		})
	}

	return mux, nil
}
