package main

import (
	"bytes"
	"log/slog"
	"os"
	"strings"
	"testing"

	approvals "github.com/approvals/go-approval-tests"
	"github.com/josephwoodward/log-ring/ringhandler"
)

func Test_FlushesAtConfiguredLevel(t *testing.T) {
	// arrange
	var buf bytes.Buffer
	logger := slog.New(ringhandler.NewHandler(
		slog.NewJSONHandler(&buf, &slog.HandlerOptions{ReplaceAttr: clearTimeAttr}),
		&ringhandler.HandlerOptions{TailSize: 10, TailLevel: slog.LevelInfo, AttrKey: "RequestId", FlushLevel: slog.LevelError},
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
	logger := slog.New(ringhandler.NewHandler(
		slog.NewJSONHandler(&buf, &slog.HandlerOptions{ReplaceAttr: clearTimeAttr, Level: slog.LevelDebug}),
		&ringhandler.HandlerOptions{TailLevel: slog.LevelInfo, FlushLevel: slog.LevelError, AttrKey: "RequestId"},
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
	//TODO: Why is this logging debug and info when the json logger below is warn
	logger := slog.New(ringhandler.NewHandler(
		slog.NewJSONHandler(&buf, &slog.HandlerOptions{ReplaceAttr: clearTimeAttr, Level: slog.LevelWarn}),
		&ringhandler.HandlerOptions{TailSize: 10, TailLevel: slog.LevelInfo, FlushLevel: slog.LevelError, AttrKey: "RequestId"},
	))

	logger.Debug("Debug 1", "RequestId", "123")
	logger.Debug("Debug 2", "RequestId", "123")

	logger.Debug("Debug 1", "RequestId", "456")
	logger.Debug("Debug 2", "RequestId", "456")

	logger.Info("Info 1", "RequestId", "123")
	logger.Info("Info 2", "RequestId", "123")

	logger.Warn("Warning 1", "RequestId", "123")
	logger.Error("Error 1", "RequestId", "123")

	actual := buf.String()
	approvals.VerifyString(t, actual)
}

func Test_LoggerFlushesOnWarn(t *testing.T) {
	// arrange
	var buf bytes.Buffer
	logger := slog.New(ringhandler.NewHandler(
		slog.NewJSONHandler(&buf, &slog.HandlerOptions{ReplaceAttr: clearTimeAttr, Level: slog.LevelInfo}),
		&ringhandler.HandlerOptions{TailSize: 10, TailLevel: slog.LevelInfo, FlushLevel: slog.LevelWarn, AttrKey: "RequestId"},
	))

	logger.Info("Info 1", "RequestId", "123")
	logger.Info("Info 2", "RequestId", "123")
	logger.Warn("Warning 1", "RequestId", "123")

	approvals.VerifyString(t, buf.String())
}

func TestMain(m *testing.M) {
	approvals.UseFolder("testdata")
	os.Exit(m.Run())
}

var clearTimeAttr = func(_ []string, a slog.Attr) slog.Attr {
	if a.Key == "time" {
		return slog.String("time", "<datetime>")
	}
	return a
}
