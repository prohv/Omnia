package unit

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"omnia/internal/logger"
)

func TestSetupLoggerModes(t *testing.T) {
	tests := []struct {
		name         string
		mode         string
		logFunc      func(l *slog.Logger)
		wantContains string
		wantNot      string
	}{
		{
			name: "verbose logs debug messages",
			mode: "verbose",
			logFunc: func(l *slog.Logger) {
				l.Debug("debug message test")
			},
			wantContains: "debug message test",
		},
		{
			name: "default ignores debug messages",
			mode: "default",
			logFunc: func(l *slog.Logger) {
				l.Debug("debug hidden")
				l.Info("info visible")
			},
			wantContains: "info visible",
			wantNot:      "debug hidden",
		},
		{
			name: "quiet logs only error messages",
			mode: "quiet",
			logFunc: func(l *slog.Logger) {
				l.Info("info hidden")
				l.Error("error visible")
			},
			wantContains: "error visible",
			wantNot:      "info hidden",
		},
		{
			name: "json outputs JSON formatted log",
			mode: "json",
			logFunc: func(l *slog.Logger) {
				l.Info("json message test")
			},
			wantContains: `"msg":"json message test"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			l := logger.SetupLogger(tt.mode, &buf)
			tt.logFunc(l)

			output := buf.String()
			if tt.wantContains != "" && !strings.Contains(output, tt.wantContains) {
				t.Errorf("SetupLogger(%s) output = %q, want containing %q", tt.mode, output, tt.wantContains)
			}
			if tt.wantNot != "" && strings.Contains(output, tt.wantNot) {
				t.Errorf("SetupLogger(%s) output = %q, should NOT contain %q", tt.mode, output, tt.wantNot)
			}
		})
	}
}
