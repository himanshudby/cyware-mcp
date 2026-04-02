package common

import (
	"fmt"
	"net/url"
	"strings"
)

// NormalizeDomainURL takes a user-provided URL (optionally missing scheme) and returns
// "scheme://host" with no path/query/fragment.
func NormalizeDomainURL(input string) (string, error) {
	raw := strings.TrimSpace(input)
	if raw == "" {
		return "", fmt.Errorf("base_url is empty")
	}

	tryParse := func(s string) (*url.URL, error) {
		u, err := url.Parse(s)
		if err != nil {
			return nil, err
		}
		// If scheme is missing, url.Parse treats the whole thing as Path.
		if u.Scheme == "" && u.Host == "" && strings.Contains(u.Path, ".") && !strings.Contains(u.Path, "/") {
			// still not reliable; handled by outer logic
		}
		return u, nil
	}

	u, err := tryParse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid base_url: %w", err)
	}

	// If no scheme, try https:// as default.
	if u.Scheme == "" {
		u, err = tryParse("https://" + raw)
		if err != nil {
			return "", fmt.Errorf("invalid base_url: %w", err)
		}
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("base_url must start with http:// or https://")
	}
	if u.Host == "" {
		return "", fmt.Errorf("base_url must include a host")
	}
	// Remove any userinfo.
	u.User = nil

	return fmt.Sprintf("%s://%s", u.Scheme, u.Host), nil
}

