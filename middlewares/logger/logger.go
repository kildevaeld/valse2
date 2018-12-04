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

			entry := log.With(zap.String("request", string(ctx.Request().URL.String())),
				zap.String("method", ctx.Request().Method),
				zap.String("remote", ctx.Request().RemoteAddr))

			if reqID := ctx.Request().Header.Get("X-Request-Id"); reqID != "" {
				entry = entry.With(zap.String("request_id", string(reqID)))
			}

			entry.Debug("started handling request")
			if err := next(ctx); err != nil {
				return err
			}

			latency := time.Since(start)

			entry.Debug("completed handling request",
				zap.Int("status", ctx.StatusCode()),
				zap.String("text_status", strong.StatusText(ctx.StatusCode())),
				zap.Duration("took", latency),
				zap.Int64("measure#.latency", latency.Nanoseconds()))

			return nil
		}
	}
}
