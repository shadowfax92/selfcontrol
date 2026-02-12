package hosts

import (
	"fmt"
	"os"
	"strings"
)

const (
	beginMarker = "# BEGIN SC BLOCK"
	endMarker   = "# END SC BLOCK"
	hostsPath   = "/etc/hosts"
)

var legacyMarkers = [][2]string{
	{"# ---- BEGIN SC BLOCK ----", "# ---- END SC BLOCK ----"},
}

func Apply(domains []string, unblocked map[string]bool, blockSubdomains bool) (bool, error) {
	content, err := os.ReadFile(hostsPath)
	if err != nil {
		return false, fmt.Errorf("reading hosts file: %w", err)
	}

	original := string(content)
	cleaned := stripLegacyBlocks(original)
	before, after := splitAroundMarkers(cleaned)

	var groups []string
	for _, d := range domains {
		if unblocked[d] {
			continue
		}
		var entry []string
		entry = append(entry, fmt.Sprintf("0.0.0.0 %s", d))
		entry = append(entry, fmt.Sprintf("::      %s", d))
		if blockSubdomains {
			entry = append(entry, fmt.Sprintf("0.0.0.0 www.%s", d))
			entry = append(entry, fmt.Sprintf("::      www.%s", d))
		}
		groups = append(groups, strings.Join(entry, "\n"))
	}

	var block string
	if len(groups) > 0 {
		block = beginMarker + "\n" + strings.Join(groups, "\n\n") + "\n" + endMarker + "\n"
	}

	newContent := before + block + after

	if newContent == original {
		return false, nil
	}

	info, err := os.Stat(hostsPath)
	if err != nil {
		return false, err
	}

	tmp := hostsPath + ".sc.tmp"
	if err := os.WriteFile(tmp, []byte(newContent), info.Mode()); err != nil {
		return false, fmt.Errorf("writing temp hosts: %w", err)
	}
	if err := os.Rename(tmp, hostsPath); err != nil {
		os.Remove(tmp)
		return false, fmt.Errorf("renaming hosts: %w", err)
	}

	return true, nil
}

func Remove() error {
	content, err := os.ReadFile(hostsPath)
	if err != nil {
		return err
	}

	original := string(content)
	cleaned := stripLegacyBlocks(original)
	before, after := splitAroundMarkers(cleaned)
	newContent := before + after

	if newContent == original {
		return nil
	}

	info, err := os.Stat(hostsPath)
	if err != nil {
		return err
	}

	return os.WriteFile(hostsPath, []byte(newContent), info.Mode())
}

func stripLegacyBlocks(content string) string {
	for _, pair := range legacyMarkers {
		beginIdx := strings.Index(content, pair[0])
		if beginIdx == -1 {
			continue
		}
		endIdx := strings.Index(content, pair[1])
		if endIdx == -1 {
			continue
		}
		tail := content[endIdx+len(pair[1]):]
		if len(tail) > 0 && tail[0] == '\n' {
			tail = tail[1:]
		}
		content = content[:beginIdx] + tail
	}
	return content
}

func splitAroundMarkers(content string) (before, after string) {
	beginIdx := strings.Index(content, beginMarker)
	if beginIdx == -1 {
		return content, ""
	}

	endIdx := strings.Index(content, endMarker)
	if endIdx == -1 {
		return content[:beginIdx], ""
	}

	after = content[endIdx+len(endMarker):]
	if len(after) > 0 && after[0] == '\n' {
		after = after[1:]
	}

	return content[:beginIdx], after
}
