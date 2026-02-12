package daemon

import (
	"context"
	"sync"
	"time"

	"sc/internal/config"
	"sc/internal/dns"
	"sc/internal/hosts"
	"sc/internal/ipc"
	"sc/internal/logs"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"

	"os"
)

type UnblockEntry struct {
	Until   time.Time `yaml:"until"`
	Started time.Time `yaml:"started"`
}

type State struct {
	Unblocked map[string]UnblockEntry `yaml:"unblocked"`
}

type Daemon struct {
	cfg       *config.Config
	cfgPath   string
	state     *State
	logger    zerolog.Logger
	mu        sync.RWMutex
	startTime time.Time
}

func New(cfg *config.Config, cfgPath string, logger zerolog.Logger) *Daemon {
	return &Daemon{
		cfg:       cfg,
		cfgPath:   cfgPath,
		state:     &State{Unblocked: make(map[string]UnblockEntry)},
		logger:    logger,
		startTime: time.Now(),
	}
}

func (d *Daemon) Run(ctx context.Context) error {
	d.loadState()
	d.tick()

	interval := d.cfg.Settings.CheckInterval.Duration
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	d.logger.Info().
		Dur("check_interval", interval).
		Dur("default_duration", d.cfg.Settings.DefaultDuration.Duration).
		Int("domains", len(d.cfg.Domains)).
		Msg("daemon started")

	for {
		select {
		case <-ctx.Done():
			d.logger.Info().Msg("daemon stopping")
			return nil
		case <-ticker.C:
			d.tick()
		}
	}
}

func (d *Daemon) tick() {
	d.mu.Lock()
	defer d.mu.Unlock()

	changed := false
	now := time.Now()

	for domain, entry := range d.state.Unblocked {
		if now.After(entry.Until) {
			delete(d.state.Unblocked, domain)
			changed = true
			d.logger.Info().Str("domain", domain).Msg("timer expired, reblocking")
			logs.Append(config.LogsPath(), logs.Entry{
				Timestamp: now,
				Event:     "reblock",
				Domain:    domain,
				Reason:    "timer_expired",
			})
		}
	}

	unblocked := make(map[string]bool)
	for domain := range d.state.Unblocked {
		unblocked[domain] = true
	}

	hostsChanged, err := hosts.Apply(d.cfg.Domains, unblocked, d.cfg.Settings.BlockSubdomains)
	if err != nil {
		d.logger.Error().Err(err).Msg("failed to apply hosts")
		return
	}

	if hostsChanged && d.cfg.Settings.FlushDNS {
		if err := dns.Flush(); err != nil {
			d.logger.Warn().Err(err).Msg("failed to flush DNS")
		}
	}

	if changed {
		d.saveState()
	}
}

func (d *Daemon) Unblock(domains []string, duration time.Duration) ipc.UnblockData {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	until := now.Add(duration)

	for _, domain := range domains {
		d.state.Unblocked[domain] = UnblockEntry{Until: until, Started: now}
		logs.Append(config.LogsPath(), logs.Entry{
			Timestamp: now,
			Event:     "unblock",
			Domain:    domain,
			Duration:  duration.String(),
		})
		d.logger.Info().Str("domain", domain).Dur("duration", duration).Msg("unblocked")
	}

	d.applyAndFlush()
	d.saveState()

	return ipc.UnblockData{Domains: domains, Duration: duration.String()}
}

func (d *Daemon) Reblock(domains []string) ipc.ReblockData {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	var reblocked []string

	if len(domains) == 0 {
		for domain := range d.state.Unblocked {
			reblocked = append(reblocked, domain)
		}
		d.state.Unblocked = make(map[string]UnblockEntry)
	} else {
		for _, domain := range domains {
			if _, ok := d.state.Unblocked[domain]; ok {
				delete(d.state.Unblocked, domain)
				reblocked = append(reblocked, domain)
			}
		}
	}

	for _, domain := range reblocked {
		logs.Append(config.LogsPath(), logs.Entry{
			Timestamp: now,
			Event:     "reblock",
			Domain:    domain,
			Reason:    "manual",
		})
		d.logger.Info().Str("domain", domain).Msg("manually reblocked")
	}

	d.applyAndFlush()
	d.saveState()

	return ipc.ReblockData{Domains: reblocked}
}

func (d *Daemon) AddDomains(domains []string) ipc.MutateData {
	d.mu.Lock()
	defer d.mu.Unlock()

	var added []string
	for _, domain := range domains {
		if d.cfg.AddDomain(domain) {
			added = append(added, domain)
		}
	}

	if len(added) > 0 {
		config.Save(d.cfg, d.cfgPath)
		d.applyAndFlush()
	}

	return ipc.MutateData{Added: added, Domains: d.cfg.Domains}
}

func (d *Daemon) RemoveDomains(domains []string) ipc.MutateData {
	d.mu.Lock()
	defer d.mu.Unlock()

	var removed []string
	for _, domain := range domains {
		if d.cfg.RemoveDomain(domain) {
			removed = append(removed, domain)
			delete(d.state.Unblocked, domain)
		}
	}

	if len(removed) > 0 {
		config.Save(d.cfg, d.cfgPath)
		d.applyAndFlush()
		d.saveState()
	}

	return ipc.MutateData{Removed: removed, Domains: d.cfg.Domains}
}

func (d *Daemon) Status() ipc.StatusData {
	d.mu.RLock()
	defer d.mu.RUnlock()

	now := time.Now()
	var entries []ipc.StatusEntry

	for _, domain := range d.cfg.Domains {
		entry := ipc.StatusEntry{Domain: domain}
		if ub, ok := d.state.Unblocked[domain]; ok {
			remaining := ub.Until.Sub(now)
			if remaining > 0 {
				entry.State = "unblocked"
				entry.Remaining = remaining.Round(time.Second).String()
			} else {
				entry.State = "blocked"
			}
		} else {
			entry.State = "blocked"
		}
		entries = append(entries, entry)
	}

	return ipc.StatusData{
		Uptime:  time.Since(d.startTime).Round(time.Second).String(),
		Domains: entries,
	}
}

func (d *Daemon) ListDomains() ipc.ListData {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return ipc.ListData{Domains: d.cfg.Domains}
}

func (d *Daemon) applyAndFlush() {
	unblocked := make(map[string]bool)
	for domain := range d.state.Unblocked {
		unblocked[domain] = true
	}

	changed, err := hosts.Apply(d.cfg.Domains, unblocked, d.cfg.Settings.BlockSubdomains)
	if err != nil {
		d.logger.Error().Err(err).Msg("failed to apply hosts")
		return
	}
	if changed && d.cfg.Settings.FlushDNS {
		if err := dns.Flush(); err != nil {
			d.logger.Warn().Err(err).Msg("failed to flush DNS")
		}
	}
}

func (d *Daemon) loadState() {
	data, err := os.ReadFile(config.StatePath())
	if err != nil {
		return
	}

	var state State
	if err := yaml.Unmarshal(data, &state); err != nil {
		d.logger.Warn().Err(err).Msg("failed to parse state file")
		return
	}

	if state.Unblocked == nil {
		state.Unblocked = make(map[string]UnblockEntry)
	}

	// Expire past-due timers
	now := time.Now()
	for domain, entry := range state.Unblocked {
		if now.After(entry.Until) {
			delete(state.Unblocked, domain)
			d.logger.Info().Str("domain", domain).Msg("expired stale unblock on startup")
		}
	}

	d.state = &state
}

func (d *Daemon) saveState() {
	data, err := yaml.Marshal(d.state)
	if err != nil {
		d.logger.Error().Err(err).Msg("failed to marshal state")
		return
	}

	if err := os.MkdirAll(config.DataDir(), 0755); err != nil {
		d.logger.Error().Err(err).Msg("failed to create data dir")
		return
	}

	tmp := config.StatePath() + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		d.logger.Error().Err(err).Msg("failed to write state")
		return
	}
	if err := os.Rename(tmp, config.StatePath()); err != nil {
		d.logger.Error().Err(err).Msg("failed to rename state")
	}
}
