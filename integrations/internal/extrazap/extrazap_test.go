package extrazap

import (
	"context"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestAttachExtract(t *testing.T) {
	tests := []struct {
		name   string
		attach bool
		want   *zap.Logger
	}{
		{
			name:   "happy_path",
			attach: true,
			want:   zap.Must(zap.NewProduction()),
		},
		{
			name:   "global_logger",
			attach: false,
			want:   zap.Must(zap.NewDevelopment()),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			globalLogger := zap.L()
			zap.ReplaceGlobals(zap.Must(zap.NewDevelopment()))

			ctx := context.Background()
			if tt.attach {
				ctx = AttachLoggerToContext(zap.Must(zap.NewProduction()), ctx)
			}
			got := ExtractLoggerFromContext(ctx)

			// Cleanup (restore original global logger) before test.
			zap.ReplaceGlobals(globalLogger)

			if g, w := zapcore.LevelOf(got.Core()), zapcore.LevelOf(tt.want.Core()); g != w {
				t.Errorf("ExtractLoggerFromContext(ctx).Level() = %v, want %v", g, w)
			}
		})
	}
}
