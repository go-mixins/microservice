package http

import (
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/go-mixins/log"

	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
	"go.opencensus.io/zpages"
)

// Checker служит для подключения внешних проверок на живость
type Checker interface {
	CheckHealth() error
}

// Replaceable functions
var (
	NowFunc = time.Now
)

// WithMetrics обвязывает http.Handler для отдачи метрик
func WithMetrics(src http.Handler, metrics http.Handler) http.Handler {
	mux := http.NewServeMux()
	zpages.Handle(mux, "/debug")
	if metrics != nil {
		mux.Handle("/metrics", metrics)
	}
	if src != nil {
		mux.Handle("/", src)
	}
	return mux
}

// clientIP implements a best effort algorithm to return the real client IP, it parses
// X-Real-IP and X-Forwarded-For in order to work properly with reverse-proxies such us: nginx or haproxy.
// This is almost unmodified code from Gin framework and all credits and my deepest thanks go to Gin developers.
func clientIP(r *http.Request) string {
	clientIP := strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if len(clientIP) > 0 {
		return clientIP
	}
	clientIP = r.Header.Get("X-Forwarded-For")
	if index := strings.IndexByte(clientIP, ','); index >= 0 {
		clientIP = clientIP[0:index]
	}
	clientIP = strings.TrimSpace(clientIP)
	if len(clientIP) > 0 {
		return clientIP
	}
	return strings.TrimSpace(r.RemoteAddr)
}

var httpOnce sync.Once

// WithLog обвязывает http.Handler для логирования запросов
func WithLog(src http.Handler, logger log.ContextLogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := r.URL.Path
		switch {
		case strings.HasPrefix(r.URL.Path, "/metrics"):
			fallthrough
		case strings.HasPrefix(r.URL.Path, "/healthz"):
			fallthrough
		case strings.HasPrefix(r.URL.Path, "/debug"):
			src.ServeHTTP(w, r)
			return
		}
		ts := NowFunc()
		ctx := r.Context()
		traceID := trace.FromContext(ctx).SpanContext().TraceID.String()
		ochttp.SetRoute(ctx, route)
		logger = logger.WithContext(log.M{
			"http_route": route,
			"client_ip":  clientIP(r),
			"trace_id":   traceID,
		})
		ctx = log.With(ctx, logger)
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, "internal server error", 500)
				logger.Errorf("%+v", err)
				logger.Debugf("panic trace: %s", debug.Stack())
				return
			}
			logger.Debugf("finished request in %v", NowFunc().Sub(ts))
		}()
		w.Header().Set("X-Trace-ID", traceID)
		src.ServeHTTP(w, r.WithContext(ctx))
	})
}
