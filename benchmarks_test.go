package logtrap_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/josephwoodward/logtrap"
)

var table = []struct {
	name string
	h    *slog.Logger
}{
	{
		name: "default handler",
		h:    slog.New(defaultHandler()),
	},
	{
		name: "ring buffer handler",
		h:    slog.New(slogHandler()),
	},
}

func slogHandler() slog.Handler {
	handler := logtrap.NewHandler(
		defaultHandler(),
		&logtrap.HandlerOptions{TailSize: 5, TailLevel: slog.LevelWarn, AttrKey: "request_id", FlushLevel: slog.LevelError},
	)
	return handler
}

func defaultHandler() slog.Handler {
	handler := slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelInfo})
	return handler
}

func Benchmark_RequestScopedLogs(b *testing.B) {
	ctx := context.Background()
	msg := "hello world!"
	attr := "request_id"

	for _, v := range table {
		b.Run(v.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				v.h.InfoContext(ctx, msg, attr, i)
			}
		})
	}

	b.ReportAllocs()
}
