# LogTrap

Save logging quota and focus on the logs that matter with LogTrap, a `log/slog` handler for Go that buffers logs in a ring buffer and only flushes them when an error or warning occurs.


### Features:

- Zero dependencies
- Write efficient buffering via a ring buffer
- Flushes logs only when the configured "flush level" is met (usually an `Error` or `Warn`)
- Helps reduce log noise and optimise logging costs
- Seamless integration with Go's `log/slog` package
- Can be scoped to attritues such as trace IDs, correlation IDs or Request IDs

### Installation

```sh
go get github.com/josephwoodward/logtrap
```

### Example Configurations

Write `Error` and `Warning` logs, but only flush `Info` and `Debug` logs when an `Error` occurs:

```go
inner := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
h := logtrap.NewHandler(inner, &logtrap.HandlerOptions{
	TailLevel:  slog.LevelInfo,
})
logger := slog.New(h)

logger.Debug("Not logged until error is encountered")
logger.Info("Not logged until error is encountered")
logger.Warn("Will log")
logger.Error("Will log and flush Info and Debug logs")
```

Write `Error` logs, and flush `Warn`, `Info` and `Debug` logs when an `Error` occurs:

```go
inner := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
h := logtrap.NewHandler(inner, &logtrap.HandlerOptions{
	TailLevel:  slog.LevelWarn,
})
logger := slog.New(h)

logger.Debug("Not logged until error is encountered")
logger.Info("Not logged until error is encountered")
logger.Warn("Not logged until error is encountered")
logger.Error("Will log and flush Info and Debug logs")
```
