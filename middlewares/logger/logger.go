package logger

import (
	"time"

	"github.com/kildevaeld/valse2/httpcontext"

	"github.com/kildevaeld/strong"
	"go.uber.org/zap"
)

func Logger() httpcontext.MiddlewareHandler {
	return LoggerWithZap(zap.L())
}

func LoggerWithZap(log *zap.Logger) httpcontext.MiddlewareHandler {
	return func(next httpcontext.HandlerFunc) httpcontext.HandlerFunc {
		return func(ctx *httpcontext.Context) error {
			start := time.Now()

			req := ctx.Request()

			entry := log.With(zap.String("request", req.URL.String()),
				zap.String("method", req.Method),
				zap.String("remote", req.RemoteAddr))

			if reqID := req.Header.Get("X-Request-Id"); reqID != "" {
				entry = entry.With(zap.String("request_id", reqID))
			}

			entry.Info("started handling request")
			if err := next(ctx); err != nil {
				entry.Info("request failed", zap.Error(err))
				return err
			}

			latency := time.Since(start)

			status := ctx.StatusCode()
			hasBody := ctx.Body() != nil
			if status == 0 {
				if hasBody {
					status = strong.StatusOK
				} else {
					status = strong.StatusNotFound
				}
			}

			entry.Info("completed handling request",
				zap.Int("status", status),
				zap.String("text_status", strong.StatusText(ctx.StatusCode())),
				zap.Duration("took", latency),
				zap.Int64("measure#.latency", latency.Nanoseconds()))

			return nil
		}
	}
}
