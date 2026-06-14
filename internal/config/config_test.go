package config_test

import (
	"testing"

	"github.com/Krushna-B/talon/internal/config"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr bool
		want    config.Config
	}{
		{
			name:    "missing DATABASE_URL",
			env:     map[string]string{"DATABASE_URL": "", "ADMIN_TOKEN": "secret"},
			wantErr: true,
		},
		{
			name: "missing ADMIN_TOKEN",
			env: map[string]string{
				"DATABASE_URL": "postgres://test",
				"ADMIN_TOKEN":  "",
			},
			wantErr: true,
		},
		{
			name: "invalid mode",
			env: map[string]string{
				"DATABASE_URL": "postgres://test",
				"ADMIN_TOKEN":  "secret",
				"MODE":         "banana",
			},
			wantErr: true,
		},
		{
			name: "defaults applied",
			env: map[string]string{
				"DATABASE_URL": "postgres://test",
				"ADMIN_TOKEN":  "secret",
				"MODE":         "",
				"HTTP_ADDR":    "",
			},
			want: config.Config{
				DatabaseURL:   "postgres://test",
				Mode:          config.ModePaper,
				HTTPAddr:      ":8080",
				AdminToken:    "secret",
				KalshiBaseURL: "https://external-api.demo.kalshi.co/trade-api/v2",
			},
		},
		{
			name: "explicit values win",
			env: map[string]string{
				"DATABASE_URL":    "postgres://test",
				"ADMIN_TOKEN":     "secret",
				"MODE":            "live",
				"HTTP_ADDR":       ":9999",
				"KALSHI_BASE_URL": "https://example.test/v2",
			},
			want: config.Config{
				DatabaseURL:   "postgres://test",
				Mode:          config.ModeLive,
				HTTPAddr:      ":9999",
				AdminToken:    "secret",
				KalshiBaseURL: "https://example.test/v2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			got, err := config.Load()

			if tt.wantErr {
				if err == nil {
					t.Fatalf("Load() succeeded, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Load() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
