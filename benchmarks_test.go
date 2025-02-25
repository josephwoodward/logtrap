package logring_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	logring "github.com/josephwoodward/log-ring"
)

var table = []struct {
	name  string
	h     *slog.Logger
	input int
}{
	{
		name:  "default handler",
		h:     slog.New(defaultHandler()),
		input: 100,
	},
	{
		name:  "ring buffer handler",
		h:     slog.New(slogHandler()),
		input: 100,
	},
}

func slogHandler() slog.Handler {
	handler := logring.NewHandler(
		defaultHandler(),
		&logring.HandlerOptions{TailSize: 5, TailLevel: slog.LevelWarn, AttrKey: "request_id", FlushLevel: slog.LevelError},
	)
	return handler
}

func defaultHandler() slog.Handler {
	handler := slog.NewTextHandler(io.Discard, &slog.HandlerOptions{ReplaceAttr: clearTimeAttr, Level: slog.LevelInfo})
	return handler
}

func Test_Benchmark(t *testing.T) {
	logger := slog.New(slogHandler())
	ctx := context.WithValue(context.Background(), "request_id", "1234")
	msg := "hello world!"

	logger.InfoContext(ctx, msg)
	logger.WarnContext(ctx, msg)
	logger.ErrorContext(ctx, msg)
}

func BenchmarkPrimeNumbers(b *testing.B) {
	ctx := context.WithValue(context.Background(), "request_id", "1234")
	msg := "hello world!"
	for _, v := range table {
		b.Run(v.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				v.h.InfoContext(ctx, msg)
				v.h.WarnContext(ctx, msg)
				v.h.WarnContext(ctx, msg)
				v.h.InfoContext(ctx, msg)
				v.h.WarnContext(ctx, msg)
				v.h.InfoContext(ctx, msg)
				v.h.ErrorContext(ctx, msg)
			}
		})
	}
	b.ReportAllocs()
}

// func BenchmarkBubbleSort(b *testing.B) {
// 	h := logring.NewHandler(
// 		slog.NewTextHandler(io.Discard, &slog.HandlerOptions{ReplaceAttr: clearTimeAttr, Level: slog.LevelDebug}),
// 		&logring.HandlerOptions{TailSize: 10, TailLevel: slog.LevelInfo, AttrKey: "request_id", FlushLevel: slog.LevelError},
// 	)
// 	logger := slog.New(h)
// 	ctx := context.WithValue(context.Background(), "request_id", "1234")
// 	for i := 0; i < b.N; i++ {
// 		logSomething(ctx, logger)
// 	}
// 	b.ReportAllocs()
// }

func logSomething(ctx context.Context, logger *slog.Logger) {
	slog.InfoContext(ctx, "Hello world!")
}
