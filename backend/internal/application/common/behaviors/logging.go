package behaviors

import (
	"context"
	"log"
	"reflect"
	"time"

	"kuorum/backend/internal/application/common/mediator"
)

// LoggingBehavior logs execution metrics for every Command and Query dispatched through mediator.
func LoggingBehavior() mediator.PipelineBehavior {
	return func(ctx context.Context, req any, next func(ctx context.Context, req any) (any, error)) (any, error) {
		reqName := reflect.TypeOf(req).String()
		startTime := time.Now()

		log.Printf("[Mediator] Executing %s", reqName)

		res, err := next(ctx, req)
		latency := time.Since(startTime)

		if err != nil {
			log.Printf("[Mediator] ERROR in %s (latency: %v): %v", reqName, latency, err)
		} else {
			log.Printf("[Mediator] SUCCESS in %s (latency: %v)", reqName, latency)
		}

		return res, err
	}
}
