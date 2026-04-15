package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// Entry represents a single audit log entry for a diff operation.
type Entry struct {
	Timestamp   time.Time         `json:"timestamp"`
	Environment string            `json:"environment"`
	Path        string            `json:"path"`
	VersionA    int               `json:"version_a"`
	VersionB    int               `json:"version_b"`
	Changes     map[string]string `json:"changes"`
	User        string            `json:"user,omitempty"`
}

// Logger writes audit entries to an output stream.
type Logger struct {
	w      io.Writer
	format string
}

// NewLogger creates a new audit Logger writing to w.
// format must be "json" or "text".
func NewLogger(w io.Writer, format string) (*Logger, error) {
	if format != "json" && format != "text" {
		return nil, fmt.Errorf("unsupported audit format %q: must be json or text", format)
	}
	return &Logger{w: w, format: format}, nil
}

// Write records an audit entry.
func (l *Logger) Write(e Entry) error {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}
	switch l.format {
	case "json":
		return l.writeJSON(e)
	default:
		return l.writeText(e)
	}
}

func (l *Logger) writeJSON(e Entry) error {
	enc := json.NewEncoder(l.w)
	enc.SetEscapeHTML(false)
	return enc.Encode(e)
}

func (l *Logger) writeText(e Entry) error {
	_, err := fmt.Fprintf(
		l.w,
		"[%s] env=%s path=%s v%d..v%d changes=%d\n",
		e.Timestamp.Format(time.RFC3339),
		e.Environment,
		e.Path,
		e.VersionA,
		e.VersionB,
		len(e.Changes),
	)
	return err
}
