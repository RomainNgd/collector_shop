package logger

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func newTestLogger() (*Logger, *bytes.Buffer, *bytes.Buffer) {
	var infoBuf, errBuf bytes.Buffer
	return &Logger{
		infoLogger:  log.New(&infoBuf, "INFO: ", 0),
		errorLogger: log.New(&errBuf, "ERROR: ", 0),
		warnLogger:  log.New(&infoBuf, "WARN: ", 0),
	}, &infoBuf, &errBuf
}

func TestLoggerInfo(t *testing.T) {
	l, infoBuf, _ := newTestLogger()

	l.Info("hello %s", "world")

	if got := infoBuf.String(); !strings.Contains(got, "INFO: hello world") {
		t.Fatalf("expected info message, got %q", got)
	}
}

func TestLoggerError(t *testing.T) {
	l, _, errBuf := newTestLogger()

	l.Error("failure %d", 42)

	if got := errBuf.String(); !strings.Contains(got, "ERROR: failure 42") {
		t.Fatalf("expected error message, got %q", got)
	}
}

func TestLoggerWarn(t *testing.T) {
	l, infoBuf, _ := newTestLogger()

	l.Warn("careful %s", "now")

	if got := infoBuf.String(); !strings.Contains(got, "WARN: careful now") {
		t.Fatalf("expected warn message, got %q", got)
	}
}

func TestNewConfiguresDistinctWriters(t *testing.T) {
	l := New()

	if l.infoLogger == nil || l.errorLogger == nil || l.warnLogger == nil {
		t.Fatal("expected all three loggers to be initialized")
	}
}

func TestDefaultLoggerIsInitialized(t *testing.T) {
	if defaultLogger == nil {
		t.Fatal("expected package init to set up defaultLogger")
	}
}

func TestGlobalHelpersDoNotPanic(t *testing.T) {
	// The global helpers write to the real default logger (stdout/stderr);
	// this only guards that they delegate correctly without panicking.
	Info("global info %s", "ok")
	Error("global error %s", "ok")
	Warn("global warn %s", "ok")
}
