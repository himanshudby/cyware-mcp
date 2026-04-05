package common

import "testing"

func TestNormalizeAuthType(t *testing.T) {
	t.Parallel()
	if got := NormalizeAuthType(Auth{Type: "OpenAPICreds"}); got != "openapicreds" {
		t.Fatalf("preserve explicit: got %q", got)
	}
	if got := NormalizeAuthType(Auth{AccessID: "a", SecretKey: "s"}); got != "openapicreds" {
		t.Fatalf("infer openapicreds: got %q", got)
	}
	if got := NormalizeAuthType(Auth{Token: "t"}); got != "token" {
		t.Fatalf("infer token: got %q", got)
	}
	if got := NormalizeAuthType(Auth{Username: "u", Password: "p"}); got != "basic" {
		t.Fatalf("infer basic: got %q", got)
	}
	if got := NormalizeAuthType(Auth{}); got != "" {
		t.Fatalf("empty: got %q", got)
	}
}
