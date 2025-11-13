package logging

import (
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Configure sets up zerolog with sane defaults.
func Configure() {
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
}

// RequestLogger provides structured request logging middleware.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		duration := time.Since(start)
		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", ww.Status()).
			Dur("duration", duration).
			Int("bytes", ww.BytesWritten()).
			Msg("http_request")
	})
}
