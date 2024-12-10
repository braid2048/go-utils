package recovery

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	labelMethod = "method"
	labelPath   = "path"
)

var (
	panicTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "panic_total",
		Help: "The total number of panics recovered",
	}, []string{labelMethod, labelPath})
)

// WithRecover recover from panic and increment the panic counter metric.
func WithRecover(fn func()) {
	defer func() {
		if err := recover(); err != nil {
			panicTotal.With(prometheus.Labels{}).Inc()
			log.Error().Msgf("Recovered from panic: %s", err)
		}
	}()

	fn()
}

// HTTPMiddleware is a Gin middleware that recovers from panics and logs the error.
func HTTPMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				panicTotal.With(prometheus.Labels{labelMethod: c.Request.Method, labelPath: c.FullPath()}).Inc()
				zerolog.Ctx(c.Request.Context()).Error().Stack().
					Err(errors.WithStack(fmt.Errorf("%+v", err))).Msg("panic recovered") //nolint:err113
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()

		c.Next()
	}
}
