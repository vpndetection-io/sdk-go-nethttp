// Package vpndetectionhttp is the official net/http middleware for the
// VPNDetection API.
//
// It classifies the visitor behind each request and puts the answer on the
// request context, where your handlers can read it. Blocking is opt-in.
//
// Everything that takes a func(http.Handler) http.Handler works with this - the
// standard library, chi, gorilla/mux, echo's WrapMiddleware. Gin has its own
// handler type and its own package, sdk-go-gin.
package vpndetectionhttp

import (
	"context"
	"encoding/json"
	"net"
	"net/http"

	vpndetection "github.com/vpndetection-io/sdk-go"
	"github.com/vpndetection-io/sdk-go/middleware"
)

// Options configure the middleware. Everything is optional except that you
// almost certainly want an APIKey: the free allowance is counted per source
// address, and a server is one source address.
type Options struct {
	middleware.Options[*http.Request]

	// OnBlocked answers a blocked request. Defaults to 403 with a short JSON
	// body. Whatever you pass must write a response.
	OnBlocked func(w http.ResponseWriter, r *http.Request, lookup *middleware.Lookup)

	// OnError answers a request the middleware could not evaluate at all,
	// which means a misconfiguration rather than a failed lookup - a condition
	// naming a member your plan does not serve, with OnMissingField set to
	// error. Defaults to 500. A failed LOOKUP never reaches this: it lets the
	// request through with the reason on the context.
	OnError func(w http.ResponseWriter, r *http.Request, err error)
}

type contextKey struct{}

// New builds the middleware, or refuses a condition that could never be what
// anyone meant.
func New(options Options) (func(http.Handler) http.Handler, error) {
	core, err := middleware.New(options.Options, DefaultIPSelector)
	if err != nil {
		return nil, err
	}
	onBlocked := options.OnBlocked
	if onBlocked == nil {
		onBlocked = refuse
	}
	onError := options.OnError
	if onError == nil {
		onError = fail
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lookup, err := core.Evaluate(r.Context(), r)
			if err != nil {
				onError(w, r, err)
				return
			}
			if lookup == nil {
				next.ServeHTTP(w, r)
				return
			}
			r = r.WithContext(context.WithValue(r.Context(), contextKey{}, lookup))
			if lookup.Blocked {
				onBlocked(w, r, lookup)
				return
			}
			next.ServeHTTP(w, r)
		})
	}, nil
}

// Must is New for a package-level variable or an init, panicking on the
// misconfiguration New would have returned.
func Must(options Options) func(http.Handler) http.Handler {
	handler, err := New(options)
	if err != nil {
		panic(err)
	}
	return handler
}

// FromContext returns what the middleware found out about this visitor, or nil
// when it has not run for this route or Skip claimed the request.
func FromContext(ctx context.Context) *middleware.Lookup {
	lookup, _ := ctx.Value(contextKey{}).(*middleware.Lookup)
	return lookup
}

var selectors = middleware.BindSelectors(func(r *http.Request) middleware.RequestView {
	return middleware.RequestView{
		Header:      r.Header.Get,
		FrameworkIP: func() string { return peerIP(r) },
	}
})

// DefaultIPSelector is the socket peer, from r.RemoteAddr, which is the only
// address net/http knows about.
//
// The standard library has no trusted-proxy setting, so behind a load balancer
// this is the balancer - a datacenter address, which a hosting rule would block
// every visitor for. If you are behind one, use HeaderIPSelector or
// XFFIPSelector instead.
var DefaultIPSelector = selectors.Default

// XFFIPSelector reads X-Forwarded-For.
//
// The LEFT-MOST entry (depth 0) is whatever the caller sent, because proxies
// append to this header, so a visitor who sets it themselves appears first and
// this returns their forgery. It is only trustworthy when an edge you control
// overwrites the header. When you know how many proxies sit in front, count
// from the right: depth 1 is the address your nearest proxy saw.
var XFFIPSelector = selectors.XFF

// HeaderIPSelector reads a single-value header your edge writes -
// HeaderIPSelector("CF-Connecting-IP") behind Cloudflare,
// HeaderIPSelector("True-Client-IP") behind Akamai. Falls back to the socket
// peer when the header is absent.
var HeaderIPSelector = selectors.Header

// RemoteAddr carries a port and, for IPv6, brackets. Stripping them here means
// a caller never sees "203.0.113.7:54321" reach a lookup and fail as malformed.
func peerIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func refuse(w http.ResponseWriter, _ *http.Request, _ *middleware.Lookup) {
	writeJSON(w, http.StatusForbidden, map[string]string{"error": "access denied"})
}

func fail(w http.ResponseWriter, _ *http.Request, _ error) {
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// Re-exported so a caller writing a condition or reading a result never has to
// import two packages.
type (
	BlockCondition = middleware.BlockCondition
	Conditions     = middleware.Conditions
	Bound          = middleware.Bound
	Lookup         = middleware.Lookup
	Result         = vpndetection.Result
)

// Gte bounds a numeric member at or above v. Gte(5).Lt(100) is a range.
func Gte(v float64) Bound { return middleware.Gte(v) }

// Gt bounds a numeric member strictly above v.
func Gt(v float64) Bound { return middleware.Gt(v) }

// Lte bounds a numeric member at or below v.
func Lte(v float64) Bound { return middleware.Lte(v) }

// Lt bounds a numeric member strictly below v.
func Lt(v float64) Bound { return middleware.Lt(v) }

// AnyOf matches a member equal to any one of these.
func AnyOf(values ...any) any { return middleware.AnyOf(values...) }
