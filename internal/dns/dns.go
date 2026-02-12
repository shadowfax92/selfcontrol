package dns

import "os/exec"

func Flush() error {
	// OS-level DNS cache — covers different macOS versions
	exec.Command("dscacheutil", "-flushcache").Run()
	exec.Command("killall", "-HUP", "mDNSResponder").Run()
	exec.Command("killall", "mDNSResponderHelper").Run()

	return nil
}
