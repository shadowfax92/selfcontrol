package ipc

const (
	CmdStatus  = "status"
	CmdUnblock = "unblock"
	CmdReblock = "reblock"
	CmdAdd     = "add"
	CmdRemove  = "remove"
	CmdList    = "list"
)

type Request struct {
	Command string            `json:"command"`
	Args    map[string]string `json:"args,omitempty"`
}

type Response struct {
	OK    bool        `json:"ok"`
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

type StatusEntry struct {
	Domain    string `json:"domain"`
	State     string `json:"state"`
	Remaining string `json:"remaining,omitempty"`
}

type StatusData struct {
	Uptime  string        `json:"uptime"`
	Domains []StatusEntry `json:"domains"`
}

type UnblockData struct {
	Domains  []string `json:"domains"`
	Duration string   `json:"duration"`
}

type ReblockData struct {
	Domains []string `json:"domains"`
}

type MutateData struct {
	Added   []string `json:"added,omitempty"`
	Removed []string `json:"removed,omitempty"`
	Domains []string `json:"domains"`
}

type ListData struct {
	Domains []string `json:"domains"`
}
