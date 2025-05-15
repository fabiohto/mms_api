package model_test

import (
	"testing"
	"time"

	"mms_api/internal/domain/model"
)

func TestIsValidPair(t *testing.T) {
	tests := []struct {
		name string
		pair string
		want bool
	}{
		{
			name: "deve retornar true para par BRLBTC",
			pair: "BRLBTC",
			want: true,
		},
		{
			name: "deve retornar true para par BRLETH",
			pair: "BRLETH",
			want: true,
		},
		{
			name: "deve retornar false para par inválido",
			pair: "INVALID",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := model.IsValidPair(tt.pair); got != tt.want {
				t.Errorf("IsValidPair() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidPeriod(t *testing.T) {
	tests := []struct {
		name   string
		period int
		want   bool
	}{
		{
			name:   "deve retornar true para período 20",
			period: model.Period20,
			want:   true,
		},
		{
			name:   "deve retornar true para período 50",
			period: model.Period50,
			want:   true,
		},
		{
			name:   "deve retornar true para período 200",
			period: model.Period200,
			want:   true,
		},
		{
			name:   "deve retornar false para período inválido",
			period: 30,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := model.IsValidPeriod(tt.period); got != tt.want {
				t.Errorf("IsValidPeriod() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMMS(t *testing.T) {
	now := time.Now()
	mms := model.MMS{
		Pair:      "BRLBTC",
		Timestamp: now,
		MMS20:     50000.0,
		MMS50:     48000.0,
		MMS200:    45000.0,
	}

	t.Run("deve criar MMS com valores corretos", func(t *testing.T) {
		if mms.Pair != "BRLBTC" {
			t.Errorf("Pair = %v, want %v", mms.Pair, "BRLBTC")
		}
		if !mms.Timestamp.Equal(now) {
			t.Errorf("Timestamp = %v, want %v", mms.Timestamp, now)
		}
		if mms.MMS20 != 50000.0 {
			t.Errorf("MMS20 = %v, want %v", mms.MMS20, 50000.0)
		}
	})
}

func TestCandle(t *testing.T) {
	now := time.Now()
	candle := model.Candle{
		Pair:      "BRLBTC",
		Timestamp: now,
		Open:      45000.0,
		High:      46000.0,
		Low:       44000.0,
		Close:     45500.0,
		Volume:    1.5,
	}

	t.Run("deve criar Candle com valores corretos", func(t *testing.T) {
		if candle.Pair != "BRLBTC" {
			t.Errorf("Pair = %v, want %v", candle.Pair, "BRLBTC")
		}
		if !candle.Timestamp.Equal(now) {
			t.Errorf("Timestamp = %v, want %v", candle.Timestamp, now)
		}
		if candle.Open != 45000.0 {
			t.Errorf("Open = %v, want %v", candle.Open, 45000.0)
		}
		if candle.High <= candle.Low {
			t.Errorf("High (%v) deve ser maior que Low (%v)", candle.High, candle.Low)
		}
	})
}
