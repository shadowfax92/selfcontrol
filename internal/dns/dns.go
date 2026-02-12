package dns

import "os/exec"

func Flush() error {
	if err := exec.Command("dscacheutil", "-flushcache").Run(); err != nil {
		return err
	}
	return exec.Command("killall", "-HUP", "mDNSResponder").Run()
}
