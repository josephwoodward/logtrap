package main

import (
	"bytes"
	"log/slog"
	"os"
	"testing"

	approvals "github.com/approvals/go-approval-tests"
	"github.com/josephwoodward/log-ring/ringhandler"
)

var (
	defaultOpts = ringhandler.HandlerOptions{TailSize: 10, TailLevel: slog.LevelInfo, AttrKey: "RequestId"}
)

func Test_LoggerOnlyLogsWarn(t *testing.T) {
	// arrange
	var buf bytes.Buffer
	logger := slog.New(ringhandler.NewHandler(
		slog.NewJSONHandler(&buf, &slog.HandlerOptions{ReplaceAttr: clearTimeAttr, Level: slog.LevelInfo}),
		&ringhandler.HandlerOptions{TailSize: 10, TailLevel: slog.LevelInfo, AttrKey: "RequestId"},
	))

	// act
	logger.Debug("Debug 1", "RequestId", "123")
	logger.Info("Info 2", "RequestId", "123")
	logger.Warn("Warning 3", "RequestId", "123")

	// assert
	approvals.VerifyString(t, buf.String())
}

func Test_LoggerFlushesDebugAndInfoLogsOnError(t *testing.T) {
	// arrange
	var buf bytes.Buffer
	logger := slog.New(ringhandler.NewHandler(
		slog.NewJSONHandler(&buf, &slog.HandlerOptions{ReplaceAttr: clearTimeAttr, Level: slog.LevelInfo}),
		&ringhandler.HandlerOptions{TailSize: 10, TailLevel: slog.LevelInfo, FlushLevel: slog.LevelError, AttrKey: "RequestId"},
	))

	logger.Debug("Debug 1", "RequestId", "123")
	logger.Debug("Debug 2", "RequestId", "123")

	logger.Debug("Debug 1", "RequestId", "456")
	logger.Debug("Debug 2", "RequestId", "456")

	logger.Info("Info 1", "RequestId", "123")
	logger.Info("Info 2", "RequestId", "123")

	logger.Warn("Warning 1", "RequestId", "123")
	logger.Error("Boom!", "RequestId", "123")

	approvals.VerifyString(t, buf.String())
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
