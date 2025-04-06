package logtrap

import (
	"container/ring"
	"context"
	"log/slog"
	"runtime"
	"sync"
	"weak"
)

type HandlerOptions struct {
	// TailSize configures the number of logs buffered that will be flushed in the event of the [logtrap.FlushLevel] being reached.
	// Default: 10.
	TailSize int

	// AttrKey is used to to index logs on [slog.Attr]
	AttrKey string

	// TailLevel configures the logs to be captured in LogTrap's buffer. TailLevel logs and lower will not be written unless the [logtrap.FlushLevel] is reached.
	// Default: slog.LevelInfo
	TailLevel slog.Leveler

	// FlushLevel determines what level to flush the buffer of log lines.
	// Default: slog.LevelError
	FlushLevel slog.Leveler
}

type LogTrapHandler struct {
	inner  slog.Handler
	opts   HandlerOptions
	mu     sync.Mutex
	buffer map[any]weak.Pointer[logs]
	goas   []groupOrAttrs
}

// groupOrAttrs holds either a group name or a list of slog.Attrs.
type groupOrAttrs struct {
	group string      // group name if non-empty
	attrs []slog.Attr // attrs if non-empty
}

type logs struct {
	*ring.Ring
	sync.RWMutex
	key any
}

// NewHandler creates a [LogTrapHandler] that writes to the handler.
// If opts is nil, the default options are used.
func NewHandler(handler slog.Handler, opts *HandlerOptions) *LogTrapHandler {
	if opts == nil {
		opts = &HandlerOptions{FlushLevel: slog.LevelError, TailSize: 10}
	}

	if opts.FlushLevel == nil {
		opts.FlushLevel = slog.LevelError
	}

	if opts.TailLevel == nil {
		opts.TailLevel = slog.LevelInfo
	}

	if opts.TailSize <= 0 {
		opts.TailSize = 10
	}

	return &LogTrapHandler{
		inner:  handler,
		buffer: make(map[any]weak.Pointer[logs]),
		opts:   *opts,
	}
}

// Enabled implements slog.Handler.
func (h *LogTrapHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

// WithAttrs implements slog.Handler.
func (h *LogTrapHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	return h.withGroupOrAttrs(groupOrAttrs{attrs: attrs})
}

// WithGroup implements slog.Handler.
func (h *LogTrapHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	return h.withGroupOrAttrs(groupOrAttrs{group: name})
}

func (h *LogTrapHandler) withGroupOrAttrs(goa groupOrAttrs) *LogTrapHandler {
	h2 := *h
	h2.goas = make([]groupOrAttrs, len(h.goas)+1)
	copy(h2.goas, h.goas)
	h2.goas[len(h2.goas)-1] = goa
	return &h2
}

func (h *LogTrapHandler) Handle(ctx context.Context, record slog.Record) error {
	if !h.inner.Enabled(ctx, record.Level) {
		return nil
	}

	if h.opts.TailSize == 0 {
		return h.inner.Handle(ctx, record)
	}

	// look for h.opts.AttrKey, context is priority followed by log attributes.
	// set a default key incase they one is not specified then handler uses same map mechanism regardless
	var key any = "nokey"
	if v := ctx.Value(h.opts.AttrKey); v != nil {
		key = v
	} else {
		// TODO: Can we use Value in map, or can we use Unique?
		record.Attrs(func(a slog.Attr) bool {
			if a.Key == h.opts.AttrKey {
				key = a.Value.Any()
				return false
			}
			return true
		})
	}

	switch {

	// flush buffer on flush level
	case record.Level >= h.opts.FlushLevel.Level():

		h.mu.Lock()
		defer h.mu.Unlock()

		var buf *logs
		b, ok := h.buffer[key]
		if ok {
			buf.key = key
			buf = b.Value()
			if buf != nil {
				var err error
				// iterate through buffer, flushing output to underlying handler
				buf.RLock()
				buf.Do(func(v any) {
					if r, ok := v.(slog.Record); ok {
						if err = h.inner.Handle(ctx, r); err != nil {
							return
						}
					}
				})
				buf.RUnlock()
				if err != nil {
					return err
				}
				delete(h.buffer, key)
			}
		}

		return h.inner.Handle(ctx, record)

	// append buffer on everything else
	default:
		// no need to capture log in buffer
		if record.Level > h.opts.TailLevel.Level() {
			return h.inner.Handle(ctx, record)
		}

		h.mu.Lock()
		defer h.mu.Unlock()

		// append to buffer, checking weak pointer
		var buf *logs
		b, ok := h.buffer[key]
		if ok {
			buf = b.Value()
			if buf != nil {
				buf.Lock()
				buf.Value = record.Clone()
				buf.Ring = buf.Next()
				h.buffer[key] = weak.Make(buf)
				buf.Unlock()
				return nil
			}
		}

		// no buffer, possibly been cleaned up so create new one
		buf = &logs{key: key}
		r := ring.New(h.opts.TailSize)
		r.Value = record.Clone()
		r = r.Next()
		buf.Ring = r

		h.buffer[key] = weak.Make(buf)

		runtime.AddCleanup(buf, func(k any) {
			h.mu.Lock()
			delete(h.buffer, k)
			h.mu.Unlock()
		}, key)

		return nil
	}
}
