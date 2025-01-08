package measure

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"
)

const (
	FieldElapsed       = "elapsed"
	FieldMeasurementID = "measurement_id"
)

var (
	logMeasurePairCounter = atomic.Uint64{}
	measureThreshold      = time.Second
)

func ContextMeasure(ctx context.Context, level slog.Level, msg string, args ...any) func() {
	then := time.Now()

	return func() {
		if elapsed := time.Since(then); elapsed >= measureThreshold {
			args = append(args, FieldElapsed, elapsed)
			slog.Log(ctx, level, msg, args...)
		}
	}
}

func Measure(level slog.Level, msg string, args ...any) func() {
	then := time.Now()

	return func() {
		if elapsed := time.Since(then); elapsed >= measureThreshold {
			args = append(args, FieldElapsed, elapsed)
			slog.Log(context.TODO(), level, msg, args...)
		}
	}
}

func ContextLogAndMeasure(ctx context.Context, level slog.Level, msg string, args ...any) func() {
	var (
		pairID = logMeasurePairCounter.Add(1)
		then   = time.Now()
	)

	args = append(args, FieldMeasurementID, pairID)
	slog.Log(ctx, level, msg, args...)

	return func() {
		if elapsed := time.Since(then); elapsed >= measureThreshold {
			args = append(args, FieldElapsed, elapsed)
			slog.Log(ctx, level, msg, args...)
		}
	}
}

func LogAndMeasure(level slog.Level, msg string, args ...any) func() {
	var (
		pairID = logMeasurePairCounter.Add(1)
		then   = time.Now()
	)

	args = append(args, FieldMeasurementID, pairID)
	slog.Log(context.TODO(), level, msg, args...)

	return func() {
		if elapsed := time.Since(then); elapsed >= measureThreshold {
			args = append(args, FieldElapsed, elapsed)
			slog.Log(context.TODO(), level, msg, args...)
		}
	}
}
