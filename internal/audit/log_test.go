package audit

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func fixedEntry() Entry {
	return Entry{
		Timestamp:   time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		Environment: "production",
		Path:        "secret/app/config",
		VersionA:    3,
		VersionB:    4,
		Changes:     map[string]string{"DB_PASS": "modified", "API_KEY": "added"},
		User:        "alice",
	}
}

func TestNewLogger_InvalidFormat(t *testing.T) {
	_, err := NewLogger(&bytes.Buffer{}, "xml")
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
}

func TestNewLogger_ValidFormats(t *testing.T) {
	for _, f := range []string{"json", "text"} {
		_, err := NewLogger(&bytes.Buffer{}, f)
		if err != nil {
			t.Fatalf("unexpected error for format %q: %v", f, err)
		}
	}
}

func TestWrite_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	l, _ := NewLogger(&buf, "json")
	e := fixedEntry()
	if err := l.Write(e); err != nil {
		t.Fatalf("Write: %v", err)
	}
	var got Entry
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if got.Environment != e.Environment {
		t.Errorf("env: got %q, want %q", got.Environment, e.Environment)
	}
	if len(got.Changes) != 2 {
		t.Errorf("changes count: got %d, want 2", len(got.Changes))
	}
}

func TestWrite_TextFormat(t *testing.T) {
	var buf bytes.Buffer
	l, _ := NewLogger(&buf, "text")
	if err := l.Write(fixedEntry()); err != nil {
		t.Fatalf("Write: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"production", "secret/app/config", "v3..v4", "changes=2"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q: %s", want, out)
		}
	}
}

func TestWrite_AutoTimestamp(t *testing.T) {
	var buf bytes.Buffer
	l, _ := NewLogger(&buf, "json")
	e := fixedEntry()
	e.Timestamp = time.Time{}
	_ = l.Write(e)
	var got Entry
	_ = json.Unmarshal(buf.Bytes(), &got)
	if got.Timestamp.IsZero() {
		t.Error("expected timestamp to be set automatically")
	}
}
