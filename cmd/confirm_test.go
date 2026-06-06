package cmd

import (
	"testing"

	"sc/internal/config"
)

func TestIsAffirmative(t *testing.T) {
	for _, s := range []string{"y", "yes", "Y", "  YES  ", "Yes\n"} {
		if !isAffirmative(s) {
			t.Errorf("isAffirmative(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"", "n", "no", "yep", "nope", "yeah", "sure"} {
		if isAffirmative(s) {
			t.Errorf("isAffirmative(%q) = true, want false", s)
		}
	}
}

func TestConfirmWeakeningSkip(t *testing.T) {
	// skip=true must short-circuit to true without touching stdin.
	if !confirmWeakening(config.Default(), true, "proceed?") {
		t.Fatal("skip=true should confirm without prompting")
	}
}

func TestConfirmWeakeningNoWarnings(t *testing.T) {
	// No warnings configured => nothing to gate, proceed without prompting.
	cfg := config.Default()
	cfg.Settings.UnblockWarnings = nil
	if !confirmWeakening(cfg, false, "proceed?") {
		t.Fatal("no warnings should confirm without prompting")
	}
}
