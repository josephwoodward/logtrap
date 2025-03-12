package logtrap

import (
	"bytes"
	"io"
	"log/slog"
	"testing"

	approvals "github.com/approvals/go-approval-tests"
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

func Test_WithLogger(t *testing.T) {
	// arrange
	var buf bytes.Buffer
	handler := NewHandler(
		slog.NewTextHandler(&buf, &slog.HandlerOptions{ReplaceAttr: clearTimeAttr, Level: slog.LevelDebug}),
		&HandlerOptions{TailSize: 10, TailLevel: slog.LevelInfo, AttrKey: "request_id", FlushLevel: slog.LevelError},
	)
	logger := slog.New(handler)

	// act
	l := logger.With(slog.String("request_id", "1234"))
	l.Debug("debug expected")
	l.Info("info expected")
	l.Error("error expected")

	// assert
	// if len(handler.buffer) != 1 {
	// 	t.Errorf("expected to find 1 map entry but found %d", len(handler.buffer))
	// }
	// actual := buf.String()
	// if !strings.Contains(actual, "debug expected") {
	// 	t.Errorf("expected to not find tailed logs but did:\n%s", actual)
	// }

	// l = logger.With(slog.String("request_id", "4321"), slog.String("request", "4321"))
	// l.Debug("debug expected")
	// l.Info("info expected")
	// l.Error("error expected")

	// assert
	// if len(handler.buffer) != 1 {
	// 	t.Errorf("expected to find 1 map entry but found %d", len(handler.buffer))
	// }

	approvals.VerifyString(t, buf.String())
}

func clearTimeAttr(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey {
		return slog.String("time", "<datetime>")
	}
	return a
}
