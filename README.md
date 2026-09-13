# [<img src="https://s3.vpndetection.io/vpndetection-public/brand/mark.svg" alt="VPNDetection" width="24"/>](https://vpndetection.io/) VPNDetection net/http Middleware

[![Go Reference](https://pkg.go.dev/badge/github.com/vpndetection-io/sdk-go-nethttp.svg)](https://pkg.go.dev/github.com/vpndetection-io/sdk-go-nethttp)
[![license](https://img.shields.io/github/license/vpndetection-io/sdk-go-nethttp.svg)](LICENSE)

The official `net/http` middleware for the [VPNDetection](https://vpndetection.io) API.

It classifies the visitor behind each request — VPN, residential proxy, Tor, hosting, CDN, relay — and puts the answer on the request context. Blocking is opt-in.

Anything that takes a `func(http.Handler) http.Handler` works with this: the standard library, [chi](https://github.com/go-chi/chi), [gorilla/mux](https://github.com/gorilla/mux), and echo through `echo.WrapMiddleware`. Gin has its own handler type and its own package, [sdk-go-gin](https://github.com/vpndetection-io/sdk-go-gin).

## Getting Started

```bash
go get github.com/vpndetection-io/sdk-go-nethttp
```

Requires Go 1.24 or newer.

You need an API key. Create one in the [console](https://app.vpndetection.io); the free tier's allowance is counted per source address, and a server is a single source address, so a key is what makes this usable in production rather than optional.

```go
package main

import (
    "fmt"
    "net/http"
    "os"

    vpndetectionhttp "github.com/vpndetection-io/sdk-go-nethttp"
    "github.com/vpndetection-io/sdk-go/middleware"
)

func main() {
    guard := vpndetectionhttp.Must(vpndetectionhttp.Options{
        Options: middleware.Options[*http.Request]{
            APIKey: os.Getenv("VPNDETECTION_API_KEY"),
        },
    })

    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        lookup := vpndetectionhttp.FromContext(r.Context())
        if lookup.Result != nil && lookup.Result.IsVpn {
            fmt.Fprintln(w, "Hello, VPN user")
            return
        }
        fmt.Fprintln(w, "Hello")
    })

    http.ListenAndServe(":8080", guard(mux))
}
```

By default nothing is blocked. Every request gets a `Lookup` on its context and your own code decides what that means — which is usually what you want, because whether a VPN visitor is a problem depends entirely on what they are doing.

`Must` panics on a misconfiguration, which suits a package-level variable. `New` returns the error instead.

## Blocking

Set a `BlockCondition` and a matching request is answered with `403` and never reaches your handlers.

```go
Options: middleware.Options[*http.Request]{
    APIKey:         os.Getenv("VPNDETECTION_API_KEY"),
    BlockCondition: middleware.Conditions{{"is_vpn": true}},
},
```

A condition is written in the shape of a result, keyed by the same names the API uses, and only the members you name are considered. That lets it reach the evidence, not just the flags:

```go
// one provider
middleware.Conditions{{"is_vpn": true, "vpn": middleware.BlockCondition{"provider": "nordvpn"}}}

// a numeric threshold
middleware.Conditions{{"resproxy": middleware.BlockCondition{"hits": vpndetectionhttp.Gte(5)}}}

// any of these
middleware.Conditions{{"vpn": middleware.BlockCondition{
    "confidence": vpndetectionhttp.AnyOf("high", "medium"),
}}}

// a list is OR
middleware.Conditions{{"is_tor": true}, {"is_resproxy": true}}
```

Values are matched by equality, strings without regard to case. `AnyOf` means any-of. `Gte`, `Gt`, `Lte` and `Lt` compare numbers and chain into a range (`Gte(5).Lt(100)`); every bound you give must hold. Members set to `false` or `nil` are ignored, so a condition states the signals you act on; one that constrains nothing would match every request, and is refused when the middleware is built rather than silently blocking all your traffic.

Replace the refusal with `OnBlocked`:

```go
OnBlocked: func(w http.ResponseWriter, r *http.Request, lookup *middleware.Lookup) {
    http.Error(w, "VPN not allowed", http.StatusForbidden)
},
```

## Where the client address comes from

This is the setting that decides whether any of the above works, and it is the one thing only you can get right.

By default the middleware uses `r.RemoteAddr`, with the port stripped. **`net/http` has no trusted-proxy setting**, so that is the socket peer and nothing else. If your app sits behind nginx, a load balancer, or a CDN, every visitor arrives wearing your proxy's address — which is a datacenter address, so a hosting rule would block all of them.

For an edge that writes the address into its own header, name the header:

```go
IPSelector: vpndetectionhttp.HeaderIPSelector("CF-Connecting-IP"),  // or True-Client-IP
```

`XFFIPSelector(0)` reads the left-most `X-Forwarded-For` entry. Be aware that the left-most entry is whatever the caller sent, because proxies append to that header — it is only trustworthy when an edge you control overwrites it. If you know how many proxies sit in front, count from the right instead: `XFFIPSelector(1)` is the address your nearest proxy saw.

Anything else, pass your own function. It receives the request and returns an address:

```go
IPSelector: func(r *http.Request) string { return r.Header.Get("X-Real-IP") },
```

If the address resolves to a private one, the middleware says so once through its logger. That is expected on localhost and is the signal to fix your configuration anywhere else.

## When a lookup fails

The request is let through, and the reason is on `Lookup.Err`. Our outage should not become yours, so a network failure, an exhausted quota or a rejected key all fail open.

```go
lookup := vpndetectionhttp.FromContext(r.Context())
if lookup.Err != nil {
    slog.Warn("vpndetection unavailable", "err", lookup.Err)
}
```

Set `FailClosed: true` to block instead. Private addresses are answered locally and never fail, so this will not lock you out in development.

## Cost and latency

Answers are cached per middleware for an hour, so a returning visitor costs nothing, and private addresses never leave the process. A cache miss is one request to our API, bounded at 2500 ms by default and not retried — on a request path, failing open quickly beats holding a visitor while we try again. Both are adjustable, and so is the cache, through a client you build yourself and pass as `Client`.

Wrap the routes that matter rather than the whole server, or skip what you do not care about:

```go
Skip: func(r *http.Request) bool { return strings.HasPrefix(r.URL.Path, "/static") },
```

If you already hold a `vpndetection.Client`, pass it as `Client` and the middleware will share it rather than building a second cache.

Beyond a few million distinct visitors a day, stop calling the API per request: [download the dataset](https://vpndetection.io/databases) and look addresses up locally instead.

## Absent is not false

Only `IP` and `IsVpn` come back on every plan. A field your plan does not include is a nil pointer, which means "not in your plan" rather than "checked, and no".

```go
vpndetection.BoolValue(lookup.Result.IsHosting)  // when you only want the flag
```

A `BlockCondition` naming a member your plan does not serve can never match, so the middleware warns once instead of failing silently. Set `OnMissingField: middleware.MissingFieldError` to make it an error.

## Other Libraries

There are official VPNDetection client libraries available for many languages including PHP, Python, Go, Java, Ruby, and many popular frameworks such as Django, Rails, and Laravel. See our GitHub at https://github.com/vpndetection-io for more.

## About VPNDetection

VPN Detection API: Accurate anonymity detection identifying VPNs, residential proxies, hosting servers, Tor nodes, CDNs, relays and more.

[<img src="https://s3.vpndetection.io/vpndetection-public/brand/mark.svg" alt="VPNDetection" width="96"/>](https://vpndetection.io/)

## License

This project is licensed under the [MIT License](LICENSE).
