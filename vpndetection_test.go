// The adapter, against a real http.Server on a real socket, with a real
// httptest.Server standing in for the API.
//
// A test server answers on the loopback, so the socket peer is a bogon and is
// answered locally without a request. Anything that needs a served answer
// therefore has to arrive wearing a public address, through a selector.

package vpndetectionhttp_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	vpndetectionhttp "github.com/vpndetection-io/sdk-go-nethttp"
	vpndetection "github.com/vpndetection-io/sdk-go/v4"
	"github.com/vpndetection-io/sdk-go/v4/middleware"
)

const publicIP = "45.83.91.1"

type answer struct {
	IP       string `json:"ip"`
	IsVpn    bool   `json:"is_vpn"`
	IsBogon  bool   `json:"is_bogon"`
	Err      string `json:"err"`
	Attached bool   `json:"attached"`
}

// A stand-in API that answers every lookup the same way and records the
// addresses it was asked about, which is what the client-address tests assert
// on rather than assume.
type fakeAPI struct {
	mu     sync.Mutex
	asked  []string
	body   string
	status int
	hang   bool
	server *httptest.Server
}

func newAPI(t *testing.T, body string) *fakeAPI {
	t.Helper()
	api := &fakeAPI{body: body}
	api.server = httptest.NewServer(http.HandlerFunc(api.serve))
	t.Cleanup(api.server.Close)
	return api
}

func (a *fakeAPI) serve(w http.ResponseWriter, r *http.Request) {
	if a.hang {
		<-r.Context().Done()
		return
	}
	ip := strings.TrimPrefix(r.URL.Path, "/")
	a.mu.Lock()
	a.asked = append(a.asked, ip)
	a.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if a.status != 0 {
		w.WriteHeader(a.status)
		fmt.Fprint(w, `{"error":"boom"}`)
		return
	}
	body := a.body
	if body == "" {
		body = `"is_vpn":true,"vpn":{"provider":"nordvpn"}`
	}
	fmt.Fprintf(w, `{"ip":%q,%s}`, ip, body)
}

func (a *fakeAPI) addresses() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]string(nil), a.asked...)
}

func (a *fakeAPI) client(t *testing.T, extra ...vpndetection.Option) *vpndetection.Client {
	t.Helper()
	options := append([]vpndetection.Option{
		vpndetection.WithBaseURL(a.server.URL),
		vpndetection.WithoutCache(),
		vpndetection.WithRetries(0),
	}, extra...)
	client, err := vpndetection.New(options...)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return client
}

func fixedIP(*http.Request) string { return publicIP }

// Stands up the middleware in front of a handler that reports what reached it.
func serve(t *testing.T, options vpndetectionhttp.Options) func(headers map[string]string) (int, answer) {
	t.Helper()
	mw, err := vpndetectionhttp.New(options)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lookup := vpndetectionhttp.FromContext(r.Context())
		out := answer{Attached: lookup != nil}
		if lookup != nil {
			out.IP = lookup.IP
			if lookup.Result != nil {
				out.IsVpn = lookup.Result.IsVpn
				out.IsBogon = lookup.Result.IsBogon
			}
			if lookup.Err != nil {
				out.Err = lookup.Err.Error()
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}))
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return func(headers map[string]string) (int, answer) {
		request, err := http.NewRequest(http.MethodGet, server.URL+"/", nil)
		if err != nil {
			t.Fatalf("NewRequest: %v", err)
		}
		for name, value := range headers {
			request.Header.Set(name, value)
		}
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatalf("Do: %v", err)
		}
		defer response.Body.Close()
		raw, _ := io.ReadAll(response.Body)
		var out answer
		// Checked, not swallowed: when a middleware writes a refusal and then
		// lets the handler run anyway, both bodies land and the result is two
		// concatenated JSON documents. Ignoring the error turns that into a
		// zero-valued answer that looks exactly like a correct refusal.
		if err := json.Unmarshal(raw, &out); err != nil {
			t.Fatalf("response body is not one JSON document (%v): %s", err, raw)
		}
		return response.StatusCode, out
	}
}

func TestEnrichesTheRequestAndLeavesTheDecisionToTheApp(t *testing.T) {
	api := newAPI(t, "")
	call := serve(t, vpndetectionhttp.Options{Options: middleware.Options[*http.Request]{
		Client: api.client(t), IPSelector: fixedIP,
	}})

	status, out := call(nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}
	if !out.Attached || !out.IsVpn || out.IP != publicIP {
		t.Errorf("answer = %+v, want an attached VPN answer for %s", out, publicIP)
	}
	if got := api.addresses(); len(got) != 1 || got[0] != publicIP {
		t.Errorf("asked about %v, want [%s]", got, publicIP)
	}
}

func TestBlocksWhenTheConditionMatches(t *testing.T) {
	api := newAPI(t, "")
	call := serve(t, vpndetectionhttp.Options{Options: middleware.Options[*http.Request]{
		Client:         api.client(t),
		IPSelector:     fixedIP,
		BlockCondition: middleware.Conditions{{"is_vpn": true}},
	}})
	if status, _ := call(nil); status != http.StatusForbidden {
		t.Errorf("status = %d, want 403", status)
	}

	clean := newAPI(t, `"is_vpn":false,"vpn":{}`)
	allowed := serve(t, vpndetectionhttp.Options{Options: middleware.Options[*http.Request]{
		Client:         clean.client(t),
		IPSelector:     fixedIP,
		BlockCondition: middleware.Conditions{{"is_vpn": true}},
	}})
	if status, _ := allowed(nil); status != http.StatusOK {
		t.Errorf("status = %d, want 200 for a clean address", status)
	}
}

// The handler must not run for a blocked request. Asserting only the status
// would pass even if it did, because the refusal is written first.
func TestABlockedRequestNeverReachesTheHandler(t *testing.T) {
	api := newAPI(t, "")
	call := serve(t, vpndetectionhttp.Options{Options: middleware.Options[*http.Request]{
		Client:         api.client(t),
		IPSelector:     fixedIP,
		BlockCondition: middleware.Conditions{{"is_vpn": true}},
	}})
	status, out := call(nil)
	if status != http.StatusForbidden || out.Attached {
		t.Errorf("status = %d, attached = %v; the handler answered a blocked request",
			status, out.Attached)
	}
}

func TestAConditionReachesTheEvidenceFields(t *testing.T) {
	nord := newAPI(t, `"is_vpn":true,"vpn":{"provider":"nordvpn"}`)
	call := serve(t, vpndetectionhttp.Options{Options: middleware.Options[*http.Request]{
		Client:     nord.client(t),
		IPSelector: fixedIP,
		BlockCondition: middleware.Conditions{
			{"vpn": middleware.BlockCondition{"provider": "mullvad"}},
		},
	}})
	if status, _ := call(nil); status != http.StatusOK {
		t.Errorf("status = %d, want 200; a different provider must not match", status)
	}

	mullvad := newAPI(t, `"is_vpn":true,"vpn":{"provider":"MULLVAD"}`)
	call2 := serve(t, vpndetectionhttp.Options{Options: middleware.Options[*http.Request]{
		Client:     mullvad.client(t),
		IPSelector: fixedIP,
		BlockCondition: middleware.Conditions{
			{"vpn": middleware.BlockCondition{"provider": "mullvad"}},
		},
	}})
	if status, _ := call2(nil); status != http.StatusForbidden {
		t.Errorf("status = %d, want 403; a provider must compare without case", status)
	}
}

func TestOnBlockedReplacesTheRefusal(t *testing.T) {
	api := newAPI(t, "")
	call := serve(t, vpndetectionhttp.Options{
		Options: middleware.Options[*http.Request]{
			Client:         api.client(t),
			IPSelector:     fixedIP,
			BlockCondition: middleware.Conditions{{"is_vpn": true}},
		},
		OnBlocked: func(w http.ResponseWriter, _ *http.Request, lookup *middleware.Lookup) {
			w.WriteHeader(http.StatusTeapot)
			fmt.Fprintf(w, `{"why":%q}`, *lookup.Result.Vpn.Provider)
		},
	})
	status, _ := call(nil)
	if status != http.StatusTeapot {
		t.Errorf("status = %d, want 418", status)
	}
}

func TestSkipLeavesTheRequestUntouched(t *testing.T) {
	api := newAPI(t, "")
	call := serve(t, vpndetectionhttp.Options{Options: middleware.Options[*http.Request]{
		Client:         api.client(t),
		IPSelector:     fixedIP,
		BlockCondition: middleware.Conditions{{"is_vpn": true}},
		Skip:           func(r *http.Request) bool { return r.URL.Path == "/" },
	}})
	status, out := call(nil)
	if status != http.StatusOK || out.Attached {
		t.Errorf("status = %d, attached = %v; skip should claim the request", status, out.Attached)
	}
	if got := api.addresses(); len(got) != 0 {
		t.Errorf("asked about %v, want nothing", got)
	}
}

func TestAFailingLookupLetsTheVisitorThrough(t *testing.T) {
	api := newAPI(t, "")
	api.status = http.StatusInternalServerError
	call := serve(t, vpndetectionhttp.Options{Options: middleware.Options[*http.Request]{
		Client:         api.client(t),
		IPSelector:     fixedIP,
		BlockCondition: middleware.Conditions{{"is_vpn": true}},
	}})
	status, out := call(nil)
	if status != http.StatusOK {
		t.Errorf("status = %d, want 200; our outage must not become the customer's", status)
	}
	if out.Err == "" {
		t.Error("the reason should be on the context")
	}
}

func TestFailClosedBlocksOnAFailedLookup(t *testing.T) {
	api := newAPI(t, "")
	api.status = http.StatusInternalServerError
	call := serve(t, vpndetectionhttp.Options{Options: middleware.Options[*http.Request]{
		Client:         api.client(t),
		IPSelector:     fixedIP,
		BlockCondition: middleware.Conditions{{"is_vpn": true}},
		FailClosed:     true,
	}})
	if status, _ := call(nil); status != http.StatusForbidden {
		t.Errorf("status = %d, want 403", status)
	}
}

func TestOnMissingFieldErrorBecomesA500(t *testing.T) {
	free := newAPI(t, `"is_vpn":true`)
	call := serve(t, vpndetectionhttp.Options{Options: middleware.Options[*http.Request]{
		Client:         free.client(t),
		IPSelector:     fixedIP,
		BlockCondition: middleware.Conditions{{"is_hosting": true}},
		OnMissingField: middleware.MissingFieldError,
	}})
	if status, _ := call(nil); status != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", status)
	}
}

// The test that matters. Every other assertion here would pass whether or not
// the selector is right, because a direct connection has nothing to confuse.
func TestAForgedXForwardedForIsIgnoredByDefault(t *testing.T) {
	forgery := map[string]string{"X-Forwarded-For": publicIP}

	plain := newAPI(t, "")
	untrusting := serve(t, vpndetectionhttp.Options{Options: middleware.Options[*http.Request]{
		Client: plain.client(t),
	}})
	_, direct := untrusting(forgery)
	if direct.IP != "127.0.0.1" {
		t.Errorf("resolved %q, want the socket peer; the header is a forgery", direct.IP)
	}
	if got := plain.addresses(); len(got) != 0 {
		t.Errorf("asked about %v; a bogon is answered locally", got)
	}

	explicit := newAPI(t, "")
	viaSelector := serve(t, vpndetectionhttp.Options{Options: middleware.Options[*http.Request]{
		Client:     explicit.client(t),
		IPSelector: vpndetectionhttp.XFFIPSelector(0),
	}})
	if _, out := viaSelector(forgery); out.IP != publicIP {
		t.Errorf("resolved %q, want %q", out.IP, publicIP)
	}
}

func TestAHeaderSelectorReadsTheEdgeThatWritesIt(t *testing.T) {
	api := newAPI(t, "")
	call := serve(t, vpndetectionhttp.Options{Options: middleware.Options[*http.Request]{
		Client:     api.client(t),
		IPSelector: vpndetectionhttp.HeaderIPSelector("CF-Connecting-IP"),
	}})
	if _, out := call(map[string]string{"CF-Connecting-IP": "45.83.91.9"}); out.IP != "45.83.91.9" {
		t.Errorf("resolved %q, want 45.83.91.9", out.IP)
	}
	if _, out := call(nil); out.IP != "127.0.0.1" {
		t.Errorf("resolved %q, want the socket peer when the edge wrote no header", out.IP)
	}
}

func TestDepthCountsTrustedHopsFromTheRight(t *testing.T) {
	api := newAPI(t, "")
	call := serve(t, vpndetectionhttp.Options{Options: middleware.Options[*http.Request]{
		Client:     api.client(t),
		IPSelector: vpndetectionhttp.XFFIPSelector(1),
	}})
	call(map[string]string{
		"X-Forwarded-For": publicIP + ", 70.41.3.18, 150.172.238.178",
	})
	if got := api.addresses(); len(got) != 1 || got[0] != "150.172.238.178" {
		t.Errorf("asked about %v, want [150.172.238.178]", got)
	}
}

// r.RemoteAddr carries a port, and an unstripped one reaches a lookup as a
// malformed address rather than as the client.
func TestTheDefaultSelectorStripsThePort(t *testing.T) {
	api := newAPI(t, "")
	call := serve(t, vpndetectionhttp.Options{Options: middleware.Options[*http.Request]{
		Client: api.client(t), IPSelector: vpndetectionhttp.DefaultIPSelector,
	}})
	if _, out := call(nil); out.IP != "127.0.0.1" {
		t.Errorf("resolved %q, want a bare 127.0.0.1", out.IP)
	}
}

func TestAPrivateClientAddressIsAnsweredLocallyAndNeverBlocks(t *testing.T) {
	api := newAPI(t, "")
	call := serve(t, vpndetectionhttp.Options{Options: middleware.Options[*http.Request]{
		Client:         api.client(t),
		BlockCondition: middleware.Conditions{{"is_vpn": true}},
	}})
	status, out := call(nil)
	if status != http.StatusOK {
		t.Errorf("status = %d; local development must not lock you out of your own app", status)
	}
	if !out.IsBogon {
		t.Error("a loopback address should be answered locally")
	}
	if got := api.addresses(); len(got) != 0 {
		t.Errorf("asked about %v, want nothing", got)
	}
}

func TestTheDeadlineBoundsTheRequest(t *testing.T) {
	api := newAPI(t, "")
	api.hang = true
	call := serve(t, vpndetectionhttp.Options{Options: middleware.Options[*http.Request]{
		Client:         api.client(t),
		IPSelector:     fixedIP,
		Timeout:        150 * time.Millisecond,
		BlockCondition: middleware.Conditions{{"is_vpn": true}},
	}})
	started := time.Now()
	status, out := call(nil)
	if status != http.StatusOK || out.Err == "" {
		t.Errorf("status = %d, err = %q; want a fail-open with the reason recorded", status, out.Err)
	}
	if elapsed := time.Since(started); elapsed > 3*time.Second {
		t.Errorf("the visitor was held %v, past the budget", elapsed)
	}
}

func TestAConditionThatConstrainsNothingIsRefused(t *testing.T) {
	for _, condition := range []middleware.Conditions{
		{{"is_vpn": false}},
		{{}},
	} {
		if _, err := vpndetectionhttp.New(vpndetectionhttp.Options{
			Options: middleware.Options[*http.Request]{BlockCondition: condition},
		}); err == nil {
			t.Errorf("New accepted %v, which would block every request", condition)
		}
	}
}
