package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
)

type contextKey string

const userIpKey contextKey = "user_ip"

func (app *application) ipFromContext(ctx context.Context) string {
	return ctx.Value(userIpKey).(string)
}

func (app *application) addIpToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ctx = context.Background()
		//	get the IP address as accurately as possible
		ip, err := getIp(r)
		if err != nil && len(ip) == 0 {
			ip = "unknown"
		}
		ctx = context.WithValue(ctx, userIpKey, ip)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getIp(r *http.Request) (string, error) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "unknown", err
	}
	userIp := net.ParseIP(ip)
	if userIp == nil {
		return "", fmt.Errorf("user IP: \"%q\" is not IP:port", r.RemoteAddr)
	}
	forward := r.Header.Get("X-Forwarded-For")
	if len(forward) > 0 {
		ip = forward
	}
	if len(ip) == 0 {
		ip = "forward"
	}

	return ip, nil
}
