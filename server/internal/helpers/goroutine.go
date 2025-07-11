package helpers

import (
	"context"
	"log/slog"

	"github.com/xjhc/alignment/server/internal/logger"
)

// GoSafe runs a function in a new goroutine with panic recovery.
// It provides standardized panic containment as mandated by ADR-008 Pattern III.
func GoSafe(ctx context.Context, fn func(ctx context.Context)) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Use structured logging for better context
				logger.GetLogger().Error("Panic recovered in goroutine",
					slog.Any("error", r),
				)
			}
		}()
		fn(ctx)
	}()
}