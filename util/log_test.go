package util

import (
	"net/http/httptest"
	"testing"
)

func TestGetClientIpFromRequest_HeaderPrecedence(t *testing.T) {
	tests := []struct {
		name   string
		headers map[string]string
		remote string
		want   string
	}{
		{
			name: "x-forwarded-for takes precedence",
			headers: map[string]string{
				"X-Forwarded-For": "198.51.100.10, 10.0.0.1",
				"X-Real-Ip":       "203.0.113.8",
			},
			remote: "127.0.0.1:12345",
			want:   "198.51.100.10",
		},
		{
			name: "x-real-ip fallback",
			headers: map[string]string{
				"X-Real-Ip": "203.0.113.8",
			},
			remote: "127.0.0.1:12345",
			want:   "203.0.113.8",
		},
		{
			name: "x-original-forwarded-for fallback",
			headers: map[string]string{
				"X-Original-Forwarded-For": "192.0.2.42, 10.0.0.2",
			},
			remote: "127.0.0.1:12345",
			want:   "192.0.2.42",
		},
		{
			name: "forwarded header fallback",
			headers: map[string]string{
				"Forwarded": "for=198.51.100.20;proto=https;by=203.0.113.43",
			},
			remote: "127.0.0.1:12345",
			want:   "198.51.100.20",
		},
		{
			name:   "remote addr fallback",
			headers: map[string]string{},
			remote: "192.0.2.99:8080",
			want:   "192.0.2.99",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://example.com", nil)
			req.RemoteAddr = tt.remote
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			if got := GetClientIpFromRequest(req); got != tt.want {
				t.Fatalf("GetClientIpFromRequest() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetForwardedHeaderIp(t *testing.T) {
	tests := []struct {
		name      string
		forwarded string
		want      string
	}{
		{
			name:      "quoted ipv6",
			forwarded: "for=\"[2001:db8:cafe::17]\";proto=https",
			want:      "2001:db8:cafe::17",
		},
		{
			name:      "multiple entries uses first for",
			forwarded: "for=198.51.100.101;proto=https, for=203.0.113.60",
			want:      "198.51.100.101",
		},
		{
			name:      "without for",
			forwarded: "proto=https;by=203.0.113.43",
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getForwardedHeaderIp(tt.forwarded); got != tt.want {
				t.Fatalf("getForwardedHeaderIp() = %q, want %q", got, tt.want)
			}
		})
	}
}
