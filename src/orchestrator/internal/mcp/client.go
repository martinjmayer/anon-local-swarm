// Package mcp implements a minimal MCP client using stdio transport.
//
// REQ-009: The orchestrator registers 10 MCP servers at startup. Each server is
// a subprocess communicating over stdin/stdout via JSON-RPC 2.0.
// If a server fails to register, the failure is logged and the orchestrator
// continues with degraded tool availability.
package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
)

// jsonRPCRequest is a JSON-RPC 2.0 request envelope.
type jsonRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int64  `json:"id,omitempty"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

// jsonRPCResponse is a JSON-RPC 2.0 response envelope.
type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ServerConfig describes a single MCP server subprocess.
type ServerConfig struct {
	Name      string
	Command   string
	Args      []string
	Env       map[string]string
	TaskTypes []string // empty = all task types
}

// CallResult holds the raw result from a tool call.
type CallResult struct {
	Content json.RawMessage
}

// server represents a running MCP server process.
type server struct {
	name    string
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  *bufio.Reader
	mu      sync.Mutex
	nextID  atomic.Int64
	healthy bool
}

// Client manages multiple MCP server subprocesses and routes tool calls.
type Client struct {
	servers map[string]*server
	mu      sync.RWMutex
}

// New creates an empty Client. Call Register to add servers.
func New() *Client {
	return &Client{servers: make(map[string]*server)}
}

// Register starts the MCP server subprocess described by cfg, performs the
// MCP initialize handshake, and registers it for tool calls.
// REQ-009: failures are non-fatal — log and continue.
func (c *Client) Register(ctx context.Context, cfg ServerConfig) error {
	cmd := exec.CommandContext(ctx, cfg.Command, cfg.Args...)

	// Inject server-specific env vars on top of the current process environment.
	cmd.Env = os.Environ()
	for k, v := range cfg.Env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("mcp %s: stdin pipe: %w", cfg.Name, err)
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("mcp %s: stdout pipe: %w", cfg.Name, err)
	}
	cmd.Stderr = os.Stderr // surface MCP server errors to swarm stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("mcp %s: start: %w", cfg.Name, err)
	}

	srv := &server{
		name:   cfg.Name,
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewReader(stdoutPipe),
	}

	// MCP initialize handshake.
	if err := srv.initialize(ctx); err != nil {
		cmd.Process.Kill() //nolint:errcheck
		return fmt.Errorf("mcp %s: initialize: %w", cfg.Name, err)
	}

	srv.healthy = true
	c.mu.Lock()
	c.servers[cfg.Name] = srv
	c.mu.Unlock()

	slog.Info("mcp: server registered", "name", cfg.Name)
	return nil
}

// Call invokes a tool on the named MCP server and returns the raw result.
// REQ-009: if the server is unavailable, returns an error (caller logs and skips).
func (c *Client) Call(ctx context.Context, serverName, toolName string, args map[string]any) (*CallResult, error) {
	c.mu.RLock()
	srv, ok := c.servers[serverName]
	c.mu.RUnlock()

	if !ok || !srv.healthy {
		return nil, fmt.Errorf("mcp: server %q not registered or unhealthy", serverName)
	}

	result, err := srv.call(ctx, "tools/call", map[string]any{
		"name":      toolName,
		"arguments": args,
	})
	if err != nil {
		return nil, fmt.Errorf("mcp %s: call %s: %w", serverName, toolName, err)
	}
	return &CallResult{Content: result}, nil
}

// Close shuts down all registered server subprocesses.
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, srv := range c.servers {
		srv.stdin.Close()
		srv.cmd.Process.Kill() //nolint:errcheck
	}
}

// HealthyServers returns the names of all healthy registered servers.
func (c *Client) HealthyServers() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var names []string
	for name, srv := range c.servers {
		if srv.healthy {
			names = append(names, name)
		}
	}
	return names
}

// initialize performs the MCP initialize + initialized handshake.
func (s *server) initialize(ctx context.Context) error {
	result, err := s.call(ctx, "initialize", map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "als-orchestrator", "version": "0.1.0"},
	})
	if err != nil {
		return err
	}
	_ = result

	// Send the initialized notification (no response expected).
	notif := jsonRPCRequest{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}
	return s.send(notif)
}

// call sends a JSON-RPC request and reads the response.
func (s *server) call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.nextID.Add(1)
	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}
	if err := s.send(req); err != nil {
		return nil, err
	}

	// Read response lines until we find one matching our request ID.
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		line, err := s.stdout.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("read response: %w", err)
		}

		var resp jsonRPCResponse
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			continue // skip non-JSON lines (server debug output, etc.)
		}
		if resp.ID != id {
			continue // not our response
		}
		if resp.Error != nil {
			return nil, fmt.Errorf("JSON-RPC error %d: %s", resp.Error.Code, resp.Error.Message)
		}
		return resp.Result, nil
	}
}

// send marshals and writes a JSON-RPC request followed by a newline.
func (s *server) send(req jsonRPCRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	data = append(data, '\n')
	_, err = s.stdin.Write(data)
	return err
}
