package logtrap

import (
	"bytes"
	"io"
	"log/slog"
	"testing"
)

func Test_DefaultHandlerOptions(t *testing.T) {
	var table = []struct {
		name string
		opts *HandlerOptions
	}{
		{name: "nil handler options", opts: nil},
		{name: "empty handler options", opts: &HandlerOptions{}},
		{name: "incorrectly configured handler options", opts: &HandlerOptions{TailSize: 0}},
	}

	for _, v := range table {
		t.Run(v.name, func(t *testing.T) {
			handler := NewHandler(
				slog.NewTextHandler(io.Discard, &slog.HandlerOptions{ReplaceAttr: clearTimeAttr, Level: slog.LevelDebug}),
				v.opts,
			)

			actual := handler.opts.FlushLevel
			if actual != slog.LevelError {
				t.Errorf("expected default FlushLevel of Error but was: %s", actual)
			}

			actual = handler.opts.TailLevel
			if actual != slog.LevelInfo {
				t.Errorf("expected default TailLevel of Info but was: %s", actual)
			}

			size := handler.opts.TailSize
			if size != 10 {
				t.Errorf("expected default TailSize of 10 but was: %d", size)
			}

			key := handler.opts.AttrKey
			if key != "" {
				t.Errorf("expected AttrKey key to be empty but was: %s", key)
			}
		})
	}
}

func Test_BufferWidth(t *testing.T) {
	var table = []struct {
		name     string
		count    int
		expected int
		opts     *HandlerOptions
	}{
		// TODO: add depth test
		{name: "nil handler options", expected: 1, count: 1, opts: nil},
		{name: "default key", expected: 1, count: 1, opts: &HandlerOptions{}},
		{name: "less than default size", expected: 5, count: 5, opts: &HandlerOptions{AttrKey: "request_id"}},
		{name: "exactly default size", expected: 10, count: 10, opts: &HandlerOptions{AttrKey: "request_id"}},
		{name: "no more than default size", expected: 10, count: 11, opts: &HandlerOptions{AttrKey: "request_id"}},
	}

	for _, v := range table {
		t.Run(v.name, func(t *testing.T) {
			var buf bytes.Buffer
			handler := NewHandler(
				slog.NewTextHandler(&buf, &slog.HandlerOptions{ReplaceAttr: clearTimeAttr, Level: slog.LevelDebug}),
				v.opts,
			)
			l := slog.New(handler)

			for i := 0; i < v.count; i++ {
				l.Info("logging debug", "request_id", i)
			}

			// assert
			got := len(handler.buffer)
			if got != v.count {
				t.Errorf("expected to find %d unique request logs but found %d", v.count, got)
			}
		})
	}
}

func clearTimeAttr(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey {
		return slog.String("time", "<datetime>")
	}
	return a
}
