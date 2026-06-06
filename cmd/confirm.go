package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"sc/internal/config"
)

// confirmWeakening is the friction gate shared by commands that weaken the
// user's blocking (unblock, remove): step through each configured warning,
// enforce a cooldown so it can't be dismissed reflexively, then ask the
// action-specific finalPrompt. Returns true only if every step is confirmed.
// skip (the -y flag) and an empty warning list both bypass it.
func confirmWeakening(cfg *config.Config, skip bool, finalPrompt string) bool {
	if skip || len(cfg.Settings.UnblockWarnings) == 0 {
		return true
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println()
	for i, w := range cfg.Settings.UnblockWarnings {
		if i > 0 {
			forceWait(30 * time.Second)
		}
		fmt.Printf("  %s [y/N] ", w)
		answer, _ := reader.ReadString('\n')
		if !isAffirmative(answer) {
			fmt.Println("Cancelled.")
			return false
		}
	}
	forceWait(30 * time.Second)

	fmt.Printf("\n  %s [y/N] ", finalPrompt)
	answer, _ := reader.ReadString('\n')
	if !isAffirmative(answer) {
		fmt.Println("Cancelled.")
		return false
	}
	return true
}

func isAffirmative(answer string) bool {
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes"
}

func forceWait(d time.Duration) {
	deadline := time.Now().Add(d)
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}
		secs := int(remaining.Round(time.Second).Seconds())
		fmt.Printf("\r  Cooling off... %ds remaining ", secs)
		time.Sleep(time.Second)
	}
	fmt.Print("\r\033[K")
}
