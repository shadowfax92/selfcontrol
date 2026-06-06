package cmd

import (
	"testing"
	"time"

	"sc/internal/config"
)

func TestResolveDemoDuration(t *testing.T) {
	cfg := config.Default()

	cfg.Settings.DemoDuration = config.Duration{}
	if got := resolveDemoDuration(cfg); got != config.DefaultDemoDuration {
		t.Fatalf("zero demo_duration => %s, want default %s", got, config.DefaultDemoDuration)
	}

	cfg.Settings.DemoDuration = config.Duration{Duration: 3 * time.Minute}
	if got := resolveDemoDuration(cfg); got != 3*time.Minute {
		t.Fatalf("configured demo_duration => %s, want 3m", got)
	}
}
