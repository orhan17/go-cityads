package middleware

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Здесь мы пишем метрики (в данном случае простая метрика)
var (
	requestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Общее количество HTTP-запросов",
		},
		[]string{"method", "endpoint"},
	)

	responseTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_time_seconds",
			Help:    "Время обработки HTTP-запросов",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)

func init() {
	prometheus.MustRegister(requestCount, responseTime)
}

func MetricsMiddleware(c *fiber.Ctx) error {
	timer := prometheus.NewTimer(responseTime.WithLabelValues(c.Method(), c.Path()))
	defer timer.ObserveDuration()

	requestCount.WithLabelValues(c.Method(), c.Path()).Inc()
	return c.Next()
}

type customResponseWriter struct {
	c *fiber.Ctx
}

func (w *customResponseWriter) Header() http.Header {
	return http.Header{}
}

func (w *customResponseWriter) Write(data []byte) (int, error) {
	return w.c.Write(data)
}

func (w *customResponseWriter) WriteHeader(statusCode int) {
	w.c.Status(statusCode)
}

func MetricsHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		w := &customResponseWriter{c: c}
		// Создаём "пустой" HTTP-запрос
		dummyReq, _ := http.NewRequest("GET", "/metrics", nil)
		promhttp.Handler().ServeHTTP(w, dummyReq)
		return nil
	}
}
