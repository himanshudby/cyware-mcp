package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cyware-labs/cyware-mcpserver/applications/co"
	"github.com/cyware-labs/cyware-mcpserver/applications/ctix"
	"github.com/cyware-labs/cyware-mcpserver/applications/general"
	"github.com/cyware-labs/cyware-mcpserver/common"
	"github.com/mark3labs/mcp-go/server"
)

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

	mcp_server_mode := cfg.Server.MCPMode
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
			Handler:           mux,
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
	}

}
