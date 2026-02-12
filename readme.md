<div align="center">

# 🧱 sc

**Block distracting websites by default. Temporarily unblock with timers.**

*The inverse of SelfControl — everything is blocked unless you explicitly allow it.*

</div>

A root daemon manages `/etc/hosts`, blocking configured domains with both IPv4 and IPv6 entries. When you need access, unblock a domain for a specific duration. When the timer expires, it's blocked again. The daemon enforces blocks every 5 seconds — you can't just edit the hosts file.

- **Default-blocked** — domains are blocked at all times unless explicitly unblocked
- **Timed unblocks** — `sc unblock reddit.com 15m` gives you 15 minutes, then reblocks
- **Instant reblock** — changed your mind? `sc reblock` puts the wall back up
- **Usage tracking** — `sc logs` shows how often you unblock and for how long

---

## Install

Requires Go 1.21+ and macOS.

```sh
git clone <repo-url> selfcontrol-go
cd selfcontrol-go
sudo make install    # builds and copies to /usr/local/bin/sc
sudo sc install      # creates launchd daemon, default config, data dirs
```

## Quick Start

```sh
# 1. Add domains to block
sc add youtube.com reddit.com x.com

# 2. Check status
sc status

# 3. Need to check something? Unblock for 10 minutes
sc unblock reddit.com 10m

# 4. Done early? Reblock immediately
sc reblock reddit.com
```

## Config

Location: `/usr/local/etc/sc/config.yaml` (created on `sc install`)

```yaml
domains:
  - youtube.com
  - reddit.com
  - x.com
  - linkedin.com
  - netflix.com

settings:
  default_duration: 15m
  check_interval: 5s
  flush_dns: true
  block_subdomains: true
```

**`domains`** — sites to block. Each gets IPv4 (`0.0.0.0`) and IPv6 (`::`) entries in `/etc/hosts`, plus `www.` variants when `block_subdomains` is enabled.

**`default_duration`** — how long `sc unblock` lasts when no duration is specified.

## CLI

```sh
sc status                     # show all domains and their state
sc unblock reddit.com 15m     # unblock for 15 minutes
sc unblock reddit.com x.com   # unblock multiple (uses default_duration)
sc reblock                    # reblock everything immediately
sc reblock reddit.com         # reblock specific domain
sc add youtube.com            # add domain to block list
sc remove youtube.com         # remove domain from block list
sc list                       # list all configured domains
sc logs                       # show unblock history and stats
sc logs --domain reddit.com   # filter logs by domain
sc logs --period today        # filter: today, week, month, all
sc version                    # print version
```

## How It Works

**Daemon** runs as root via launchd (`com.sc.daemon`). Every 5 seconds it:
1. Expires any timed unblocks that are past due
2. Rebuilds the `/etc/hosts` block section
3. Flushes macOS DNS cache if anything changed

**CLI** talks to the daemon over a unix socket at `/usr/local/var/sc/sc.sock`. The socket is world-readable so non-root users can send commands, but only the root daemon writes to `/etc/hosts`.

**Hosts file** entries sit between `# BEGIN SC BLOCK` / `# END SC BLOCK` markers. Content outside the markers is never touched.

## Paths

| What | Path |
|------|------|
| Config | `/usr/local/etc/sc/config.yaml` |
| State | `/usr/local/var/sc/state.yaml` |
| Logs | `/usr/local/var/sc/logs.jsonl` |
| Socket | `/usr/local/var/sc/sc.sock` |
| Daemon log | `/usr/local/var/sc/daemon.log` |
| Plist | `/Library/LaunchDaemons/com.sc.daemon.plist` |
| Binary | `/usr/local/bin/sc` |

## Uninstall

```sh
sudo sc uninstall    # stops daemon, removes plist, cleans /etc/hosts
```

---

> This is a personal tool I built for my own workflow. I'm sharing it in case it's useful to others, but I'm not actively seeking feature requests or contributions. Feel free to fork and adapt it to your needs.
