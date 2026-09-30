package server

import (
	"context"
	"math"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Balestrino/italian-weather-alert/internal/publicquery"
)

// PublicLimits applies one sliding request budget to both JSON and MCP.
// TrustedProxies controls whether X-Forwarded-For may influence client identity.
type PublicLimits struct {
	Allowance      int
	Window         time.Duration
	MaxPageSize    int
	TrustedProxies []netip.Prefix
}

type budgetState struct {
	Limit, Remaining, ResetSeconds, RetryAfterSeconds, MaxPageSize int
	Allowed                                                        bool
}

type budgetContextKey struct{}

type slidingBudget struct {
	mu        sync.Mutex
	limits    PublicLimits
	now       func() time.Time
	requests  map[string][]time.Time
	lastSweep time.Time
}

func newSlidingBudget(limits PublicLimits, now func() time.Time) *slidingBudget {
	if limits.Allowance < 1 || limits.Window <= 0 || limits.MaxPageSize < 1 {
		panic("invalid public request limits")
	}
	return &slidingBudget{limits: limits, now: now, requests: map[string][]time.Time{}}
}

func (b *slidingBudget) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state := b.take(clientAddress(r, b.limits.TrustedProxies))
		writeBudgetHeaders(w.Header(), state, b.limits.Window)
		if !state.Allowed {
			w.Header().Set("Retry-After", strconv.Itoa(state.RetryAfterSeconds))
			servedAt := b.now().UTC()
			response := publicResponse{Error: &publicError{Code: "rate_limited", Message: "the shared anonymous IP request budget is exhausted", RetryAfterSeconds: &state.RetryAfterSeconds, Candidates: []publicquery.Municipality{}}, ServedAt: &servedAt}
			reply(w, http.StatusTooManyRequests, response)
			return
		}
		ctx := context.WithValue(r.Context(), budgetContextKey{}, state)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (b *slidingBudget) take(client string) budgetState {
	now := b.now().UTC()
	cutoff := now.Add(-b.limits.Window)
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.lastSweep.IsZero() || now.Sub(b.lastSweep) >= b.limits.Window {
		for key, values := range b.requests {
			if len(values) == 0 || !values[len(values)-1].After(cutoff) {
				delete(b.requests, key)
			}
		}
		b.lastSweep = now
	}
	values := b.requests[client]
	first := 0
	for first < len(values) && !values[first].After(cutoff) {
		first++
	}
	values = values[first:]
	state := budgetState{Limit: b.limits.Allowance, MaxPageSize: b.limits.MaxPageSize}
	if len(values) >= b.limits.Allowance {
		state.Remaining = 0
		state.ResetSeconds = secondsUntil(now, values[0].Add(b.limits.Window))
		state.RetryAfterSeconds = state.ResetSeconds
		b.requests[client] = values
		return state
	}
	values = append(values, now)
	b.requests[client] = values
	state.Allowed = true
	state.Remaining = b.limits.Allowance - len(values)
	state.ResetSeconds = secondsUntil(now, values[0].Add(b.limits.Window))
	return state
}

func secondsUntil(now, deadline time.Time) int {
	seconds := int(math.Ceil(deadline.Sub(now).Seconds()))
	if seconds < 1 {
		return 1
	}
	return seconds
}

func writeBudgetHeaders(header http.Header, state budgetState, window time.Duration) {
	windowSeconds := int(math.Ceil(window.Seconds()))
	header.Set("RateLimit-Policy", `"iwa-public";q=`+strconv.Itoa(state.Limit)+`;w=`+strconv.Itoa(windowSeconds))
	header.Set("RateLimit", `"iwa-public";r=`+strconv.Itoa(state.Remaining)+`;t=`+strconv.Itoa(state.ResetSeconds))
	header.Set("X-IWA-Max-Page-Size", strconv.Itoa(state.MaxPageSize))
}

func clientAddress(request *http.Request, trusted []netip.Prefix) string {
	peer := parseRemoteAddress(request.RemoteAddr)
	if !isTrusted(peer, trusted) {
		return canonicalAddress(peer)
	}
	forwarded := request.Header.Values("X-Forwarded-For")
	if len(forwarded) == 0 {
		return canonicalAddress(peer)
	}
	var chain []netip.Addr
	for _, line := range forwarded {
		for _, raw := range strings.Split(line, ",") {
			address, err := netip.ParseAddr(strings.TrimSpace(raw))
			if err != nil {
				return canonicalAddress(peer)
			}
			chain = append(chain, address.Unmap())
		}
	}
	for index := len(chain) - 1; index >= 0; index-- {
		if !isTrusted(chain[index], trusted) {
			return canonicalAddress(chain[index])
		}
	}
	if len(chain) > 0 {
		return canonicalAddress(chain[0])
	}
	return canonicalAddress(peer)
}

func parseRemoteAddress(value string) netip.Addr {
	if addressPort, err := netip.ParseAddrPort(value); err == nil {
		return addressPort.Addr().Unmap()
	}
	if host, _, err := net.SplitHostPort(value); err == nil {
		if address, parseErr := netip.ParseAddr(host); parseErr == nil {
			return address.Unmap()
		}
	}
	address, _ := netip.ParseAddr(value)
	return address.Unmap()
}

func canonicalAddress(address netip.Addr) string {
	if !address.IsValid() {
		return "invalid-peer"
	}
	return address.Unmap().String()
}

func isTrusted(address netip.Addr, trusted []netip.Prefix) bool {
	if !address.IsValid() {
		return false
	}
	for _, prefix := range trusted {
		if prefix.Contains(address) {
			return true
		}
	}
	return false
}

func budgetFromContext(ctx context.Context) (budgetState, bool) {
	state, ok := ctx.Value(budgetContextKey{}).(budgetState)
	return state, ok
}
