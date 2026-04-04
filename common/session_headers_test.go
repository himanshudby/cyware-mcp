package common

import (
	"net/http"
	"os"
	"testing"
)

func TestApplySessionFromMCPHeaders_CTIXCO(t *testing.T) {
	ClearSession("sid-1")
	t.Cleanup(func() { ClearSession("sid-1") })

	h := http.Header{}
	h.Set(HeaderCywareCTIXBaseURL, "https://ctix.example.com")
	h.Set(HeaderCywareCTIXAccessID, "acc1")
	h.Set(HeaderCywareCTIXSecretKey, "sec1")
	h.Set(HeaderCywareCOBaseURL, "https://co.example.com/soar/")
	h.Set(HeaderCywareCOAccessID, "acc2")
	h.Set(HeaderCywareCOSecretKey, "sec2")

	ApplySessionFromMCPHeaders("sid-1", h)

	ctix, ok := GetSessionCTIX("sid-1")
	if !ok || ctix.BASE_URL != "https://ctix.example.com" || ctix.Auth.Type != "openapicreds" || ctix.Auth.AccessID != "acc1" || ctix.Auth.SecretKey != "sec1" {
		t.Fatalf("unexpected CTIX session: %#v ok=%v", ctix, ok)
	}
	co, ok := GetSessionCO("sid-1")
	if !ok || co.BASE_URL != "https://co.example.com/soar/" || co.Auth.AccessID != "acc2" {
		t.Fatalf("unexpected CO session: %#v ok=%v", co, ok)
	}
}

func TestApplySessionFromMCPHeaders_PartialIgnored(t *testing.T) {
	ClearSession("sid-2")
	t.Cleanup(func() { ClearSession("sid-2") })

	h := http.Header{}
	h.Set(HeaderCywareCTIXBaseURL, "https://ctix.example.com")
	// missing access + secret — must not set partial CTIX

	ApplySessionFromMCPHeaders("sid-2", h)

	if _, ok := GetSessionCTIX("sid-2"); ok {
		t.Fatal("expected no CTIX session when headers incomplete")
	}
}

func TestApplySessionFromMCPHeaders_EnvDisable(t *testing.T) {
	_ = os.Setenv("CYWARE_DISABLE_UPSTREAM_SESSION_HEADERS", "1")
	t.Cleanup(func() {
		_ = os.Unsetenv("CYWARE_DISABLE_UPSTREAM_SESSION_HEADERS")
		ClearSession("sid-3")
	})

	h := http.Header{}
	h.Set(HeaderCywareCTIXBaseURL, "https://ctix.example.com")
	h.Set(HeaderCywareCTIXAccessID, "acc1")
	h.Set(HeaderCywareCTIXSecretKey, "sec1")

	ApplySessionFromMCPHeaders("sid-3", h)

	if _, ok := GetSessionCTIX("sid-3"); ok {
		t.Fatal("expected headers ignored when CYWARE_DISABLE_UPSTREAM_SESSION_HEADERS is set")
	}
}
