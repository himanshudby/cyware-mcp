package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/cyware-labs/cyware-mcpserver/applications/co"
	"github.com/cyware-labs/cyware-mcpserver/applications/ctix"
	"github.com/cyware-labs/cyware-mcpserver/applications/general"
	"github.com/cyware-labs/cyware-mcpserver/common"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type statusCapturingWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusCapturingWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func withHTTPAccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusCapturingWriter{ResponseWriter: w, status: http.StatusOK}
		defer func() {
			log.Printf("%s %s -> %d (%s)", r.Method, r.URL.Path, sw.status, time.Since(start).Truncate(time.Millisecond))
		}()
		next.ServeHTTP(sw, r)
	})
}

func withHTTPRecover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic in HTTP handler %s %s: %v", r.Method, r.URL.Path, rec)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func withBearerAuth(next http.Handler, token string) http.Handler {
	if strings.TrimSpace(token) == "" {
		return next
	}
	expected := "Bearer " + strings.TrimSpace(token)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != expected {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type streamableHTTPSession struct {
	sessionID           string
	notificationChannel chan mcp.JSONRPCNotification
	initialized         bool
	mu                  sync.RWMutex
}

func (s *streamableHTTPSession) SessionID() string { return s.sessionID }
func (s *streamableHTTPSession) NotificationChannel() chan<- mcp.JSONRPCNotification {
	return s.notificationChannel
}
func (s *streamableHTTPSession) Initialize() {
	s.mu.Lock()
	s.initialized = true
	s.mu.Unlock()
}
func (s *streamableHTTPSession) Initialized() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.initialized
}

func randomSessionID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

// writeStreamableHTTPJSONError responds with application/json so MCP HTTP clients do not fail
// transport detection (they expect application/json or text/event-stream, not text/plain).
func writeStreamableHTTPJSONError(w http.ResponseWriter, httpStatus int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"jsonrpc": "2.0",
		"id":      nil,
		"error": map[string]any{
			"code":    -32000,
			"message": message,
		},
	})
}

func withBearerAuthStreamableHTTP(next http.Handler, token string) http.Handler {
	if strings.TrimSpace(token) == "" {
		return next
	}
	expected := "Bearer " + strings.TrimSpace(token)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != expected {
			writeStreamableHTTPJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	envPath := flag.String("config_path", "cmd/config.yaml", "Path to the .yaml file")
	flag.Parse()

	s := server.NewMCPServer(
		"CYWARE-MCP-SERVER",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
		server.WithRecovery(),
		server.WithInstructions(`
		# Cyware MCP Server 
		This server provides tools to access Cyware Products - CTIX(Cyware Threat Intel Exchange), CFTR, CSAP, CO platform functionalities and features.
		## ❗⚠️ Don't use tools where its mentioned in the tools description that it must be explicitly invoked.
		`),
	)

	cfg, err := common.Load(*envPath)

	if err != nil {
		log.Fatal(err)
	}

	ctix.Initialize(cfg, s)
	general.Initialize(s)
	co.Initialize(cfg, s)

	mcp_server_mode := normalizeMCPMode(cfg.Server.MCPMode)
	if mcp_server_mode == "" {
		mcp_server_mode = "stdio"
	}

	switch mcp_server_mode {
	case "stdio":
		if err := server.ServeStdio(s); err != nil {
			log.Fatalf("Server error: %v\n", err)
		}
	case "sse":
		host := cfg.Server.Host
		if host == "" {
			host = "127.0.0.1"
		}
		// Prefer binding to all interfaces (IPv4 + IPv6) when host is 0.0.0.0
		// so PaaS internal networks can connect over either family.
		if host == "0.0.0.0" {
			host = ""
		}
		port := cfg.Server.Port
		if port == "" {
			if envPort := strings.TrimSpace(os.Getenv("PORT")); envPort != "" {
				port = envPort
			} else {
				port = "5421"
			}
		}
		addr := net.JoinHostPort(host, port)

		sseServer := server.NewSSEServer(
			s,
			server.WithBaseURL(cfg.Server.BaseURL),
			server.WithStaticBasePath(cfg.Server.BasePath),
		)

		mux := http.NewServeMux()
		basePath := strings.TrimSuffix(cfg.Server.BasePath, "/")
		mux.Handle(basePath+sseServer.CompleteSsePath(), withBearerAuth(sseServer.SSEHandler(), cfg.Server.AuthToken))
		mux.Handle(basePath+sseServer.CompleteMessagePath(), withBearerAuth(sseServer.MessageHandler(), cfg.Server.AuthToken))

		httpServer := &http.Server{
			Addr:              addr,
			Handler:           withHTTPAccessLog(withHTTPRecover(mux)),
			ReadHeaderTimeout: 10 * time.Second,
		}

		go func() {
			log.Printf("MCP SSE listening on %s (base_path=%q)", addr, cfg.Server.BasePath)
			if strings.TrimSpace(cfg.Server.BaseURL) != "" {
				log.Printf("MCP SSE advertised base_url=%q", cfg.Server.BaseURL)
			}
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Server error: %v", err)
			}
		}()

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		<-stop
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(ctx)
	case "streamable_http", "http":
		host := cfg.Server.Host
		if host == "" {
			host = "127.0.0.1"
		}
		if host == "0.0.0.0" {
			host = ""
		}
		port := cfg.Server.Port
		if port == "" {
			if envPort := strings.TrimSpace(os.Getenv("PORT")); envPort != "" {
				port = envPort
			} else {
				port = "5421"
			}
		}
		addr := net.JoinHostPort(host, port)

		basePath := strings.TrimSuffix(cfg.Server.BasePath, "/")
		if basePath == "" {
			basePath = "/mcp"
		}

		mux := http.NewServeMux()
		var sessions sync.Map // sessionID -> *streamableHTTPSession

		// Railway (and other PaaS) health checks often hit `/`.
		// Return 200 so the service is considered healthy.
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/", "/healthz", "/health":
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("ok"))
			default:
				http.NotFound(w, r)
			}
		})

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Route guard: only handle exact basePath (and basePath/).
			if r.URL.Path != basePath && r.URL.Path != basePath+"/" {
				writeStreamableHTTPJSONError(w, http.StatusNotFound, "not found")
				return
			}

			switch r.Method {
			case http.MethodPost:
				raw, err := io.ReadAll(r.Body)
				if err != nil {
					writeStreamableHTTPJSONError(w, http.StatusBadRequest, "failed to read request body")
					return
				}
				defer r.Body.Close()

				// Peek method to decide session handling.
				var peek struct {
					Method string `json:"method"`
				}
				_ = json.Unmarshal(raw, &peek)

				sessionID := strings.TrimSpace(r.Header.Get("Mcp-Session-Id"))
				var sess *streamableHTTPSession

				if sessionID == "" {
					if peek.Method != "initialize" {
						writeStreamableHTTPJSONError(w, http.StatusBadRequest, "missing Mcp-Session-Id")
						return
					}

					sid, err := randomSessionID()
					if err != nil {
						writeStreamableHTTPJSONError(w, http.StatusInternalServerError, "failed to create session")
						return
					}
					sess = &streamableHTTPSession{
						sessionID:           sid,
						notificationChannel: make(chan mcp.JSONRPCNotification, 64),
					}
					if err := s.RegisterSession(r.Context(), sess); err != nil {
						writeStreamableHTTPJSONError(w, http.StatusInternalServerError, "failed to register session")
						return
					}
					sessions.Store(sid, sess)
					sessionID = sid
				} else {
					val, ok := sessions.Load(sessionID)
					if !ok {
						// Client transport expects 404 to mean "session terminated, re-initialize".
						writeStreamableHTTPJSONError(w, http.StatusNotFound, "session not found")
						return
					}
					sess = val.(*streamableHTTPSession)
				}

				ctx := s.WithContext(r.Context(), sess)

				resp := s.HandleMessage(ctx, raw)

				// If we created a session on initialize, return it for the client.
				if peek.Method == "initialize" && sessionID != "" {
					w.Header().Set("Mcp-Session-Id", sessionID)
				}

				if resp == nil {
					// Notification acknowledgement; clients (e.g. Java/TS MCP SDKs) expect
					// application/json, not default text/plain or an unset Content-Type.
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(http.StatusAccepted)
					return
				}
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				_ = json.NewEncoder(w).Encode(resp)
			case http.MethodDelete:
				sessionID := strings.TrimSpace(r.Header.Get("Mcp-Session-Id"))
				if sessionID != "" {
					sessions.Delete(sessionID)
					s.UnregisterSession(r.Context(), sessionID)
				}
				w.WriteHeader(http.StatusNoContent)
			default:
				writeStreamableHTTPJSONError(w, http.StatusMethodNotAllowed, "method not allowed: use POST for MCP requests")
			}
		})

		mux.Handle(basePath, withBearerAuthStreamableHTTP(handler, cfg.Server.AuthToken))
		mux.Handle(basePath+"/", withBearerAuthStreamableHTTP(handler, cfg.Server.AuthToken))

		httpServer := &http.Server{
			Addr:              addr,
			Handler:           withHTTPAccessLog(withHTTPRecover(mux)),
			ReadHeaderTimeout: 10 * time.Second,
		}

		go func() {
			log.Printf("MCP Streamable HTTP listening on %s (endpoint=%q)", addr, basePath)
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Server error: %v", err)
			}
		}()

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		<-stop
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(ctx)
	default:
		log.Fatalf(
			`invalid server.mcp_mode %q (or server.transport): use one of "stdio", "sse", "streamable_http" (aliases: "http", "streamable-http"), or set env MCP_MODE`,
			strings.TrimSpace(cfg.Server.MCPMode),
		)
	}

}

// normalizeMCPMode maps common variants to the values used in cmd/main switch.
func normalizeMCPMode(mode string) string {
	m := strings.TrimSpace(strings.ToLower(mode))
	m = strings.ReplaceAll(m, "-", "_")
	switch m {
	case "streamablehttp":
		return "streamable_http"
	default:
		return m
	}
}
