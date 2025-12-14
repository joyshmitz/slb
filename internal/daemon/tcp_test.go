package daemon

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net"
	"testing"
	"time"

	"github.com/charmbracelet/log"
)

func TestTCPServer_AuthHandshake(t *testing.T) {
	logger := log.New(io.Discard)

	srv, err := NewTCPServer(TCPServerOptions{
		Addr:        "127.0.0.1:0",
		RequireAuth: true,
		AllowedIPs:  []string{"127.0.0.1"},
		ValidateAuth: func(_ context.Context, sessionKey string) (bool, error) {
			return sessionKey == "good", nil
		},
	}, logger)
	if err != nil {
		t.Fatalf("NewTCPServer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = srv.Start(ctx) }()
	t.Cleanup(func() { _ = srv.Stop() })

	addr := srv.listener.Addr().String()

	t.Run("rejects bad auth", func(t *testing.T) {
		conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
		if err != nil {
			t.Fatalf("dial: %v", err)
		}
		defer conn.Close()

		_ = conn.SetDeadline(time.Now().Add(500 * time.Millisecond))
		_, _ = conn.Write([]byte(`{"auth":"bad"}` + "\n"))
		_, _ = conn.Write([]byte(`{"method":"ping","id":1}` + "\n"))

		r := bufio.NewReader(conn)
		if _, err := r.ReadBytes('\n'); err == nil {
			t.Fatalf("expected connection to be rejected")
		}
	})

	t.Run("accepts good auth", func(t *testing.T) {
		conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
		if err != nil {
			t.Fatalf("dial: %v", err)
		}
		defer conn.Close()

		_ = conn.SetDeadline(time.Now().Add(500 * time.Millisecond))
		if _, err := conn.Write([]byte(`{"auth":"good"}` + "\n")); err != nil {
			t.Fatalf("write handshake: %v", err)
		}
		if _, err := conn.Write([]byte(`{"method":"ping","id":1}` + "\n")); err != nil {
			t.Fatalf("write ping: %v", err)
		}

		r := bufio.NewReader(conn)
		line, err := r.ReadBytes('\n')
		if err != nil {
			t.Fatalf("read response: %v", err)
		}

		var resp RPCResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if resp.Error != nil {
			t.Fatalf("unexpected rpc error: %s", resp.Error.Message)
		}
		m, ok := resp.Result.(map[string]any)
		if !ok {
			t.Fatalf("unexpected result type %T", resp.Result)
		}
		if v, ok := m["pong"].(bool); !ok || !v {
			t.Fatalf("expected pong=true, got %v", m["pong"])
		}
	})
}

func TestTCPServer_IPAllowlist(t *testing.T) {
	logger := log.New(io.Discard)

	srv, err := NewTCPServer(TCPServerOptions{
		Addr:        "127.0.0.1:0",
		RequireAuth: false,
		AllowedIPs:  []string{"10.0.0.0/8"},
	}, logger)
	if err != nil {
		t.Fatalf("NewTCPServer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = srv.Start(ctx) }()
	t.Cleanup(func() { _ = srv.Stop() })

	addr := srv.listener.Addr().String()

	conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(500 * time.Millisecond))
	_, _ = conn.Write([]byte(`{"auth":""}` + "\n"))
	_, _ = conn.Write([]byte(`{"method":"ping","id":1}` + "\n"))

	r := bufio.NewReader(conn)
	if _, err := r.ReadBytes('\n'); err == nil {
		t.Fatalf("expected connection to be rejected by allowlist")
	}
}
