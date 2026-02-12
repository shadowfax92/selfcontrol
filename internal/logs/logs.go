package logs

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"
)

type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	Event     string    `json:"event"`
	Domain    string    `json:"domain"`
	Duration  string    `json:"duration,omitempty"`
	Reason    string    `json:"reason,omitempty"`
}

type QueryOpts struct {
	Domain string
	Period string
}

type DomainStats struct {
	Domain       string
	Unblocks     int
	TotalTime    time.Duration
	LastUnblock  time.Time
}

func Append(path string, entry Entry) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = f.Write(data)
	return err
}

func Query(path string, opts QueryOpts) ([]Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	cutoff := periodCutoff(opts.Period)

	var entries []Entry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var e Entry
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			continue
		}
		if !cutoff.IsZero() && e.Timestamp.Before(cutoff) {
			continue
		}
		if opts.Domain != "" && e.Domain != opts.Domain {
			continue
		}
		entries = append(entries, e)
	}

	return entries, scanner.Err()
}

func Stats(entries []Entry) []DomainStats {
	unblockTimes := make(map[string][]time.Time)
	reblockTimes := make(map[string][]time.Time)
	unblockCounts := make(map[string]int)

	for _, e := range entries {
		switch e.Event {
		case "unblock":
			unblockCounts[e.Domain]++
			unblockTimes[e.Domain] = append(unblockTimes[e.Domain], e.Timestamp)
		case "reblock":
			reblockTimes[e.Domain] = append(reblockTimes[e.Domain], e.Timestamp)
		}
	}

	statsMap := make(map[string]*DomainStats)
	for domain, count := range unblockCounts {
		ds := &DomainStats{
			Domain:   domain,
			Unblocks: count,
		}

		unblocksForDomain := unblockTimes[domain]
		reblocksForDomain := reblockTimes[domain]

		for i, ut := range unblocksForDomain {
			if i < len(reblocksForDomain) {
				ds.TotalTime += reblocksForDomain[i].Sub(ut)
			}
			if ds.LastUnblock.IsZero() || ut.After(ds.LastUnblock) {
				ds.LastUnblock = ut
			}
		}

		statsMap[domain] = ds
	}

	var result []DomainStats
	for _, ds := range statsMap {
		result = append(result, *ds)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Unblocks > result[j].Unblocks
	})

	return result
}

func periodCutoff(period string) time.Time {
	now := time.Now()
	switch period {
	case "today":
		y, m, d := now.Date()
		return time.Date(y, m, d, 0, 0, 0, 0, now.Location())
	case "week":
		return now.AddDate(0, 0, -7)
	case "month":
		return now.AddDate(0, 0, -30)
	default:
		return time.Time{}
	}
}

func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh%dm", h, m)
}
