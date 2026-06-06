package config

import (
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestDefaultDemoDuration(t *testing.T) {
	if DefaultDemoDuration != 5*time.Minute {
		t.Fatalf("DefaultDemoDuration = %s, want 5m", DefaultDemoDuration)
	}
	if got := Default().Settings.DemoDuration.Duration; got != 5*time.Minute {
		t.Fatalf("Default() demo duration = %s, want 5m", got)
	}
}

func TestDemoDurationRoundTrip(t *testing.T) {
	cfg := Default()
	cfg.Settings.DemoDuration = Duration{10 * time.Minute}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}

	var got Config
	if err := yaml.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Settings.DemoDuration.Duration != 10*time.Minute {
		t.Fatalf("round-trip demo duration = %s, want 10m", got.Settings.DemoDuration.Duration)
	}
}

func TestOmittedDemoDurationKeepsDefault(t *testing.T) {
	cfg := Default()
	body := []byte("settings:\n  default_duration: 20m\n")
	if err := yaml.Unmarshal(body, cfg); err != nil {
		t.Fatal(err)
	}
	if got := cfg.Settings.DemoDuration.Duration; got != 5*time.Minute {
		t.Fatalf("omitted demo_duration = %s, want 5m default preserved", got)
	}
}
