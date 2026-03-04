// Copyright 2021 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web/context"
)

func getIpInfo(clientIp string) string {
	if clientIp == "" {
		return ""
	}

	first := strings.TrimSpace(strings.Split(clientIp, ",")[0])
	if host, _, err := net.SplitHostPort(first); err == nil {
		return strings.Trim(host, "[]")
	}

	return strings.Trim(first, "[]")
}

func getForwardedHeaderIp(forwarded string) string {
	// RFC 7239 format example:
	// Forwarded: for=203.0.113.43, for="[2001:db8:cafe::17]"
	for _, part := range strings.Split(forwarded, ",") {
		for _, token := range strings.Split(part, ";") {
			token = strings.TrimSpace(token)
			lowerToken := strings.ToLower(token)
			if !strings.HasPrefix(lowerToken, "for=") {
				continue
			}

			// Extract the value after "for=" while preserving the original token's casing.
			ip := strings.TrimSpace(token[len("for="):])
			ip = strings.Trim(ip, "\"")
			return getIpInfo(ip)
		}
	}

	return ""
}

func getIpFromHeaders(req *http.Request) string {
	if clientIp := req.Header.Get("X-Forwarded-For"); clientIp != "" {
		return getIpInfo(clientIp)
	}

	if clientIp := req.Header.Get("X-Real-Ip"); clientIp != "" {
		return getIpInfo(clientIp)
	}

	if clientIp := req.Header.Get("X-Original-Forwarded-For"); clientIp != "" {
		return getIpInfo(clientIp)
	}

	if clientIp := getForwardedHeaderIp(req.Header.Get("Forwarded")); clientIp != "" {
		return clientIp
	}

	return ""
}

func GetClientIpFromRequest(req *http.Request) string {
	clientIp := getIpFromHeaders(req)
	if clientIp == "" {
		ipPort := strings.Split(req.RemoteAddr, ":")
		if len(ipPort) >= 1 && len(ipPort) <= 2 {
			clientIp = ipPort[0]
		} else if len(ipPort) > 2 {
			idx := strings.LastIndex(req.RemoteAddr, ":")
			clientIp = req.RemoteAddr[0:idx]
			clientIp = strings.TrimLeft(clientIp, "[")
			clientIp = strings.TrimRight(clientIp, "]")
		}
	}

	return getIpInfo(clientIp)
}

func LogInfo(ctx *context.Context, f string, v ...interface{}) {
	ipString := fmt.Sprintf("(%s) ", GetClientIpFromRequest(ctx.Request))
	logs.Info(ipString+f, v...)
}

func LogWarning(ctx *context.Context, f string, v ...interface{}) {
	ipString := fmt.Sprintf("(%s) ", GetClientIpFromRequest(ctx.Request))
	logs.Warning(ipString+f, v...)
}
