package common

import (
	"net/http"
	"os"
	"strings"
)

// HTTP header names for passing CTIX/CO upstream credentials on each MCP HTTP request
// (Streamable HTTP or SSE message POST). Values match configure-* tools (openapicreds).
//
// Use HTTPS in production; these headers carry secrets.
const (
	HeaderCywareCTIXBaseURL   = "X-Cyware-CTIX-Base-Url"
	HeaderCywareCTIXAccessID  = "X-Cyware-CTIX-Access-Id"
	HeaderCywareCTIXSecretKey = "X-Cyware-CTIX-Secret-Key"
	HeaderCywareCOBaseURL     = "X-Cyware-CO-Base-Url"
	HeaderCywareCOAccessID    = "X-Cyware-CO-Access-Id"
	HeaderCywareCOSecretKey   = "X-Cyware-CO-Secret-Key"
)

// ApplySessionFromMCPHeaders stores CTIX/CO Application config for this MCP session when
// the client sends the Cyware headers. Omitted products are left unchanged (YAML defaults
// or prior configure-* / header values still apply).
//
// For each product, all three headers (base URL, access id, secret key) must be non-empty;
// otherwise that product is skipped for this request.
//
// Set env CYWARE_DISABLE_UPSTREAM_SESSION_HEADERS=1 to ignore these headers (operators only).
func ApplySessionFromMCPHeaders(sessionID string, h http.Header) {
	if sessionID == "" || h == nil {
		return
	}
	if upstreamSessionHeadersDisabled() {
		return
	}

	if app, ok := ctixAppFromHeaders(h); ok {
		SetSessionCTIX(sessionID, app)
	}
	if app, ok := coAppFromHeaders(h); ok {
		SetSessionCO(sessionID, app)
	}
}

func upstreamSessionHeadersDisabled() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv("CYWARE_DISABLE_UPSTREAM_SESSION_HEADERS")))
	return v == "1" || v == "true" || v == "yes"
}

func ctixAppFromHeaders(h http.Header) (Application, bool) {
	base := strings.TrimSpace(h.Get(HeaderCywareCTIXBaseURL))
	access := strings.TrimSpace(h.Get(HeaderCywareCTIXAccessID))
	secret := strings.TrimSpace(h.Get(HeaderCywareCTIXSecretKey))
	if base == "" || access == "" || secret == "" {
		return Application{}, false
	}
	if _, err := NormalizeDomainURL(base); err != nil {
		return Application{}, false
	}
	return Application{
		BASE_URL: base,
		Auth: Auth{
			Type:      "openapicreds",
			AccessID:  access,
			SecretKey: secret,
		},
	}, true
}

func coAppFromHeaders(h http.Header) (Application, bool) {
	base := strings.TrimSpace(h.Get(HeaderCywareCOBaseURL))
	access := strings.TrimSpace(h.Get(HeaderCywareCOAccessID))
	secret := strings.TrimSpace(h.Get(HeaderCywareCOSecretKey))
	if base == "" || access == "" || secret == "" {
		return Application{}, false
	}
	if _, err := NormalizeDomainURL(base); err != nil {
		return Application{}, false
	}
	return Application{
		BASE_URL: base,
		Auth: Auth{
			Type:      "openapicreds",
			AccessID:  access,
			SecretKey: secret,
		},
	}, true
}
