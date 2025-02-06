package ringhandler

import (
	"container/ring"
	"context"
	"log/slog"
	"sync"
)

type HandlerOptions struct {
	// Number of per request logs buffered that will be flushed in the event of an error. Default is 10.
	TailSize int

	// Attribute to index logs on
	AttrKey string

	TailLevel slog.Leveler

	// FlushLevel determines what level to flush the buffer of log lines. Default is Error.
	FlushLevel slog.Leveler
}
type commonHandler struct {
	inner  slog.Handler
	opts   HandlerOptions
	mu     *sync.Mutex
	buffer map[string]*ring.Ring
}

type LogTailHandler struct {
	*commonHandler
}

// NewHandler creates a [LogTailHandler] that writes to the handler.
// If opts is nil, the default options are used.
func NewHandler(handler slog.Handler, opts *HandlerOptions) *LogTailHandler {
	if opts == nil {
		opts = &HandlerOptions{FlushLevel: slog.LevelError, TailSize: 10}
	}

	if opts.FlushLevel == nil {
		opts.FlushLevel = slog.LevelError
	}

	if opts.TailSize == 0 {
		opts.TailSize = 10
	}

	return &LogTailHandler{
		&commonHandler{
			inner:  handler,
			buffer: make(map[string]*ring.Ring),
			opts:   *opts,
			mu:     &sync.Mutex{},
		},
	}
}

// Enabled implements slog.Handler.
func (h *LogTailHandler) Enabled(ctx context.Context, level slog.Level) bool {
	h.inner.Enabled(ctx, level)
	return true
}

// WithAttrs implements slog.Handler.
func (h *LogTailHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.inner.WithAttrs(attrs)
	return h
}

// WithGroup implements slog.Handler.
func (h *LogTailHandler) WithGroup(name string) slog.Handler {
	h.WithGroup(name)
	return h
}

func (h *LogTailHandler) Handle(ctx context.Context, record slog.Record) error {
	var key string
	record.Attrs(func(a slog.Attr) bool {
		if a.Key == h.opts.AttrKey {
			key = a.Value.String()
			return false
		}
		return true
	})

	switch record.Level {

	// flush buffer on error
	case h.opts.FlushLevel:
		if buf, ok := h.buffer[key]; ok {
			var err error
			buf.Do(func(v any) {
				if r, ok := v.(slog.Record); ok {
					if err = h.inner.Handle(ctx, r); err != nil {
						return
					}
				}
			})
			if err != nil {
				return err
			}
		}

		return h.inner.Handle(ctx, record)

	// append buffer on everything else
	default:
		h.mu.Lock()
		defer h.mu.Unlock()

		// no need to capture log in buffer
		if record.Level > h.opts.TailLevel.Level() {
			return h.inner.Handle(ctx, record)
		}

		if h.opts.TailSize == 0 {
			return h.inner.Handle(ctx, record)
		}

		if buf, ok := h.buffer[key]; ok {
			buf.Value = record.Clone()
			buf = buf.Next()
			h.buffer[key] = buf
		} else {
			buf = ring.New(h.opts.TailSize)
			buf.Value = record.Clone()
			buf = buf.Next()
			h.buffer[key] = buf
		}

		return nil
	}
}
