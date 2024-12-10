package prometrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	labelMethod     = "method"
	labelStatusCode = "statusCode"
	labelPath       = "path"
)

var (
	// HTTPRequestDurationSecondsBuckets are the default buckets used for HTTP request duration metrics.
	HTTPRequestDurationSecondsBuckets = []float64{.01, .025, .05, .1, .25, .5, 1}

	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{labelMethod, labelStatusCode, labelPath},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: HTTPRequestDurationSecondsBuckets,
		},
		[]string{labelMethod, labelStatusCode, labelPath},
	)
)

// InitMetrics initializes and registers Prometheus metrics.
func InitMetrics() {
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration)
}

// Middleware returns a Gin middleware that records HTTP requests and responses.
func Middleware(skipPathsMap map[string]struct{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := skipPathsMap[c.Request.URL.Path]; ok {
			c.Next()
			return
		}

		startTime := time.Now()

		// 记录请求总数
		httpRequestsTotal.With(genLabels(c)).Inc()

		// 执行请求
		c.Next()

		httpRequestDuration.With(genLabels(c)).Observe(time.Since(startTime).Seconds())
	}
}

// PrometheusHandler returns a Prometheus HTTP handler for exposing metrics.
func PrometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// genLabels make labels values.
func genLabels(c *gin.Context) prometheus.Labels {
	labels := make(prometheus.Labels)
	labels[labelMethod] = c.Request.Method
	labels[labelStatusCode] = strconv.Itoa(c.Writer.Status())
	labels[labelPath] = c.FullPath()

	return labels
}
