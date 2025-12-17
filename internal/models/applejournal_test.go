package models_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kpod13/journal2day1/internal/models"
)

func TestCocoaTimestampToTime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		timestamp float64
		want      time.Time
	}{
		{
			name:      "zero timestamp returns epoch",
			timestamp: 0,
			want:      time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:      "known timestamp converts correctly",
			timestamp: 784043393,
			want:      time.Date(2025, 11, 5, 13, 49, 53, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := models.CocoaTimestampToTime(tt.timestamp)

			require.WithinDuration(t, tt.want, got, time.Second)
		})
	}
}
