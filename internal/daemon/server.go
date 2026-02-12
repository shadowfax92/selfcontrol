package daemon

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"sc/internal/ipc"

	"github.com/rs/zerolog"
)

type Server struct {
	daemon   *Daemon
	sockPath string
	logger   zerolog.Logger
	listener net.Listener
}

func NewServer(d *Daemon, sockPath string, logger zerolog.Logger) *Server {
	return &Server{
		daemon:   d,
		sockPath: sockPath,
		logger:   logger,
	}
}

func (s *Server) Start() error {
	if _, err := os.Stat(s.sockPath); err == nil {
		conn, err := net.DialTimeout("unix", s.sockPath, 500*time.Millisecond)
		if err == nil {
			conn.Close()
			return fmt.Errorf("daemon already running (socket %s is active)", s.sockPath)
		}
		os.Remove(s.sockPath)
	}

	ln, err := net.Listen("unix", s.sockPath)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", s.sockPath, err)
	}
	s.listener = ln

	if err := os.Chmod(s.sockPath, 0666); err != nil {
		ln.Close()
		return fmt.Errorf("chmod socket: %w", err)
	}

	go s.accept()
	s.logger.Info().Str("socket", s.sockPath).Msg("IPC server listening")
	return nil
}

func (s *Server) Stop() {
	if s.listener != nil {
		s.listener.Close()
	}
	os.Remove(s.sockPath)
}

func (s *Server) accept() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		return
	}

	var req ipc.Request
	if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
		s.writeResponse(conn, ipc.Response{Error: "invalid request"})
		return
	}

	var resp ipc.Response
	switch req.Command {
	case ipc.CmdStatus:
		resp = s.handleStatus()
	case ipc.CmdUnblock:
		resp = s.handleUnblock(req)
	case ipc.CmdReblock:
		resp = s.handleReblock(req)
	case ipc.CmdAdd:
		resp = s.handleAdd(req)
	case ipc.CmdRemove:
		resp = s.handleRemove(req)
	case ipc.CmdList:
		resp = s.handleList()
	default:
		resp = ipc.Response{Error: fmt.Sprintf("unknown command: %s", req.Command)}
	}

	s.writeResponse(conn, resp)
}

func (s *Server) handleStatus() ipc.Response {
	data := s.daemon.Status()
	return ipc.Response{OK: true, Data: data}
}

func (s *Server) handleUnblock(req ipc.Request) ipc.Response {
	durationStr := req.Args["duration"]
	if durationStr == "" {
		return ipc.Response{Error: "duration required"}
	}

	dur, err := time.ParseDuration(durationStr)
	if err != nil {
		return ipc.Response{Error: fmt.Sprintf("invalid duration: %s", durationStr)}
	}

	var domains []string
	if domainsStr := req.Args["domains"]; domainsStr != "" {
		domains = strings.Split(domainsStr, ",")
		for i := range domains {
			domains[i] = strings.TrimSpace(domains[i])
		}
		for _, d := range domains {
			if !s.daemon.cfg.HasDomain(d) {
				return ipc.Response{Error: fmt.Sprintf("domain %q not in block list", d)}
			}
		}
	} else {
		domains = s.daemon.cfg.Domains
	}

	data := s.daemon.Unblock(domains, dur)
	return ipc.Response{OK: true, Data: data}
}

func (s *Server) handleReblock(req ipc.Request) ipc.Response {
	var domains []string
	if domainsStr := req.Args["domains"]; domainsStr != "" {
		domains = strings.Split(domainsStr, ",")
		for i := range domains {
			domains[i] = strings.TrimSpace(domains[i])
		}
	}

	data := s.daemon.Reblock(domains)
	return ipc.Response{OK: true, Data: data}
}

func (s *Server) handleAdd(req ipc.Request) ipc.Response {
	domainsStr := req.Args["domains"]
	if domainsStr == "" {
		return ipc.Response{Error: "domains required"}
	}

	domains := strings.Split(domainsStr, ",")
	for i := range domains {
		domains[i] = strings.TrimSpace(domains[i])
	}

	data := s.daemon.AddDomains(domains)
	return ipc.Response{OK: true, Data: data}
}

func (s *Server) handleRemove(req ipc.Request) ipc.Response {
	domainsStr := req.Args["domains"]
	if domainsStr == "" {
		return ipc.Response{Error: "domains required"}
	}

	domains := strings.Split(domainsStr, ",")
	for i := range domains {
		domains[i] = strings.TrimSpace(domains[i])
	}

	data := s.daemon.RemoveDomains(domains)
	return ipc.Response{OK: true, Data: data}
}

func (s *Server) handleList() ipc.Response {
	data := s.daemon.ListDomains()
	return ipc.Response{OK: true, Data: data}
}

func (s *Server) writeResponse(conn net.Conn, resp ipc.Response) {
	data, _ := json.Marshal(resp)
	data = append(data, '\n')
	conn.Write(data)
}
