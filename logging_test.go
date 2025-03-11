package logtrap_test

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"

	approvals "github.com/approvals/go-approval-tests"
	"github.com/josephwoodward/logtrap"
)

func Test_GetsKeyFromContext(t *testing.T) {
	// arrange
	var buf bytes.Buffer
	logger := slog.New(logtrap.NewHandler(
		slog.NewTextHandler(&buf, &slog.HandlerOptions{ReplaceAttr: clearTimeAttr, Level: slog.LevelDebug}),
		&logtrap.HandlerOptions{TailSize: 10, TailLevel: slog.LevelInfo, AttrKey: "request_id", FlushLevel: slog.LevelError},
	))

	// act
	ctx := context.WithValue(context.Background(), "request_id", "1234")
	logger.DebugContext(ctx, "debug expected")
	logger.InfoContext(ctx, "info expected")
	logger.ErrorContext(ctx, "log error")

	// assert
	approvals.VerifyString(t, buf.String())
	// if !strings.Contains(buf.String(), "debug expected") {
	// 	t.Errorf("expected to see debug logs but didn't find it:\n%s", buf.String())
	// }
	// if !strings.Contains(buf.String(), "info expected") {
	// 	t.Errorf("expected to see info logs but didn't find it:\n%s", buf.String())
	// }
}

func Test_FlushesAtConfiguredLevel(t *testing.T) {
	// arrange
	var buf bytes.Buffer
	logger := slog.New(logtrap.NewHandler(
		slog.NewTextHandler(&buf, &slog.HandlerOptions{ReplaceAttr: clearTimeAttr, Level: slog.LevelDebug}),
		&logtrap.HandlerOptions{TailSize: 10, TailLevel: slog.LevelInfo, AttrKey: "RequestId", FlushLevel: slog.LevelError},
	))

	// act
	logger.Debug("expected", "RequestId", "123")
	logger.Error("Error", "RequestId", "123")

	// assert
	if !strings.Contains(buf.String(), "expected") {
		t.Errorf("expected to see debug logs but didn't find it:\n%s", buf.String())
	}
}

// Verifies that tailed logs won't be logged if flush level is not encountered
func Test_LoggerOnlyLogsAboveTailLevel(t *testing.T) {
	// arrange
	var buf bytes.Buffer
	// logs level info and debug, flushes on error
	logger := slog.New(logtrap.NewHandler(
		slog.NewTextHandler(&buf, &slog.HandlerOptions{ReplaceAttr: clearTimeAttr, Level: slog.LevelDebug}),
		&logtrap.HandlerOptions{TailLevel: slog.LevelInfo, FlushLevel: slog.LevelError, AttrKey: "RequestId"},
	))

	// act
	logger.Debug("should not log", "RequestId", "123")
	logger.Info("should not log", "RequestId", "123")
	logger.Warn("should log", "RequestId", "123")

	// assert
	actual := buf.String()
	if strings.Contains(actual, "should not log") {
		t.Errorf("expected to not find tailed logs but did:\n%s", actual)
	}

	if !strings.Contains(actual, "should log") {
		t.Errorf("expected to find flush log but did not:\n%s", actual)
	}
}

func Test_LoggerFlushesDebugAndInfoLogsOnError(t *testing.T) {
	// arrange
	var buf bytes.Buffer
	logger := slog.New(logtrap.NewHandler(
		slog.NewTextHandler(&buf, &slog.HandlerOptions{ReplaceAttr: clearTimeAttr, Level: slog.LevelDebug}),
		&logtrap.HandlerOptions{TailSize: 0, TailLevel: slog.LevelInfo, FlushLevel: slog.LevelError, AttrKey: "RequestId"},
	))

	// act
	logger.Debug("Debug 1", "RequestId", "123")
	logger.Debug("Debug 2", "RequestId", "123")

	logger.Debug("Should also log", "RequestId", "456")
	logger.Debug("Should also log", "RequestId", "456")

	logger.Info("Info 1", "RequestId", "123")
	logger.Info("Info 2", "RequestId", "123")

	logger.Warn("Warning 1", "RequestId", "123")
	logger.Error("Error 1", "RequestId", "123")

	// assert
	approvals.VerifyString(t, buf.String())
}

func Test_LoggerFlushesOnWarn(t *testing.T) {
	t.Run("flushes on flush level", func(t *testing.T) {
		t.Skip()
		// arrange
		var buf bytes.Buffer
		logger := slog.New(logtrap.NewHandler(
			slog.NewTextHandler(&buf, &slog.HandlerOptions{ReplaceAttr: clearTimeAttr, Level: slog.LevelInfo}),
			&logtrap.HandlerOptions{TailSize: 10, TailLevel: slog.LevelInfo, FlushLevel: slog.LevelWarn, AttrKey: "RequestId"},
		))

		// act
		logger.Info("Info 1", "RequestId", "123")
		logger.Info("Info 2", "RequestId", "123")
		logger.Warn("Warning 1", "RequestId", "123")

		// assert
		approvals.VerifyString(t, buf.String())
	})

	t.Run("flushes in levels above flush level", func(t *testing.T) {
		// arrange
		var buf bytes.Buffer
		logger := slog.New(logtrap.NewHandler(
			slog.NewTextHandler(&buf, &slog.HandlerOptions{ReplaceAttr: clearTimeAttr, Level: slog.LevelDebug}),
			&logtrap.HandlerOptions{TailSize: 10, TailLevel: slog.LevelInfo, FlushLevel: slog.LevelWarn, AttrKey: "RequestId"},
		))

		l := logger.With("request_id", 123)

		// act
		l.Info("should log")
		l.Info("should log")
		l.Error("Flush all")

		// assert
		actual := buf.String()
		if !strings.Contains(actual, "should log") {
			t.Errorf("expected to find info logs but did not: %s", actual)
		}
		if !strings.Contains(actual, "Flush all") {
			t.Errorf("expected to find flush log but did not: %s", actual)
		}
	})

}

func TestMain(m *testing.M) {
	approvals.UseFolder("testdata")
	os.Exit(m.Run())
}

func clearTimeAttr(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey {
		return slog.String("time", "<datetime>")
	}
	return a
}
