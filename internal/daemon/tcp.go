package daemon

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/charmbracelet/log"
)

// TCPServerOptions configures the optional TCP listener used for remote agents.
type TCPServerOptions struct {
	Addr        string
	RequireAuth bool
	AllowedIPs  []string

	// ValidateAuth returns true if the provided session key is authorized to connect.
	// If nil, any non-empty auth key is accepted when RequireAuth is true.
	ValidateAuth func(ctx context.Context, sessionKey string) (bool, error)
}

// NewTCPServer starts a TCP listener implementing the same line-delimited JSON-RPC protocol
// as the Unix socket, with an initial auth handshake.
//
// Handshake: client must first send a single line JSON object: {"auth":"<session_key>"}.
// If RequireAuth is true, the auth value must validate; otherwise it may be empty.
func NewTCPServer(opts TCPServerOptions, logger *log.Logger) (*IPCServer, error) {
	addr := strings.TrimSpace(opts.Addr)
	if addr == "" {
		return nil, fmt.Errorf("tcp addr is required")
	}

	allowedNets, err := parseAllowedIPNets(opts.AllowedIPs)
	if err != nil {
		return nil, err
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen tcp %s: %w", addr, err)
	}

	guard := func(conn net.Conn, scanner *bufio.Scanner) error {
		remoteIP, err := extractRemoteIP(conn.RemoteAddr())
		if err != nil {
			return err
		}
		if len(allowedNets) > 0 && !ipAllowed(remoteIP, allowedNets) {
			return fmt.Errorf("tcp client ip not allowed: %s", remoteIP.String())
		}

		// Require a handshake line from the client.
		_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		defer func() { _ = conn.SetReadDeadline(time.Time{}) }()

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("handshake read error: %w", err)
			}
			return fmt.Errorf("handshake missing")
		}

		var hello struct {
			Auth string `json:"auth"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &hello); err != nil {
			return fmt.Errorf("invalid handshake: %w", err)
		}

		auth := strings.TrimSpace(hello.Auth)
		if opts.RequireAuth && auth == "" {
			return fmt.Errorf("auth required")
		}

		if auth != "" && opts.ValidateAuth != nil {
			vctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			ok, err := opts.ValidateAuth(vctx, auth)
			if err != nil {
				return fmt.Errorf("auth validation error: %w", err)
			}
			if !ok {
				return fmt.Errorf("invalid auth")
			}
		}

		return nil
	}

	return newIPCServer(ln, addr, logger, nil, guard), nil
}

func parseAllowedIPNets(values []string) ([]*net.IPNet, error) {
	nets := make([]*net.IPNet, 0, len(values))
	for _, raw := range values {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		if strings.Contains(raw, "/") {
			_, n, err := net.ParseCIDR(raw)
			if err != nil {
				return nil, fmt.Errorf("invalid allowed ip cidr %q: %w", raw, err)
			}
			nets = append(nets, n)
			continue
		}

		ip := net.ParseIP(raw)
		if ip == nil {
			return nil, fmt.Errorf("invalid allowed ip %q", raw)
		}
		if ip.To4() != nil {
			ip4 := ip.To4()
			nets = append(nets, &net.IPNet{IP: ip4, Mask: net.CIDRMask(32, 32)})
		} else if ip16 := ip.To16(); ip16 != nil {
			nets = append(nets, &net.IPNet{IP: ip16, Mask: net.CIDRMask(128, 128)})
		} else {
			return nil, fmt.Errorf("invalid allowed ip %q", raw)
		}
	}
	return nets, nil
}

func extractRemoteIP(addr net.Addr) (net.IP, error) {
	if addr == nil {
		return nil, fmt.Errorf("missing remote address")
	}

	if tcp, ok := addr.(*net.TCPAddr); ok && tcp.IP != nil {
		return tcp.IP, nil
	}

	host, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		ip := net.ParseIP(addr.String())
		if ip == nil {
			return nil, fmt.Errorf("unable to parse remote ip: %s", addr.String())
		}
		return ip, nil
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return nil, fmt.Errorf("unable to parse remote ip: %s", host)
	}
	return ip, nil
}

func ipAllowed(ip net.IP, allowed []*net.IPNet) bool {
	if ip == nil {
		return false
	}
	for _, n := range allowed {
		if n != nil && n.Contains(ip) {
			return true
		}
	}
	return false
}
