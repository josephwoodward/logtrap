package logring

import (
	"bytes"
	"log/slog"
	"testing"

	approvals "github.com/approvals/go-approval-tests"
)

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
	if len(handler.buffer) != 1 {
		t.Errorf("expected to find 1 map entry but found %d", len(handler.buffer))
	}
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

var clearTimeAttr = func(_ []string, a slog.Attr) slog.Attr {
	if a.Key == "time" {
		return slog.String("time", "<datetime>")
	}
	return a
}
