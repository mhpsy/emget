package emby

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAuthorizationHeader_WithoutToken(t *testing.T) {
	c := NewClient("https://example.com", &http.Client{Timeout: 2 * time.Second})
	c.deviceID = "dev-1"
	c.version = "0.1.0"
	c.clientName = "emget"
	c.device = "unit-test"

	got := c.authHeader("")
	want := `MediaBrowser Client="emget", Device="unit-test", DeviceId="dev-1", Version="0.1.0"`
	if got != want {
		t.Errorf("\ngot  %s\nwant %s", got, want)
	}
}

func TestAuthorizationHeader_WithToken(t *testing.T) {
	c := NewClient("https://example.com", nil)
	c.deviceID = "dev-1"
	c.version = "0.1.0"
	c.clientName = "emget"
	c.device = "unit-test"

	got := c.authHeader("tok-xyz")
	if !strings.Contains(got, `Token="tok-xyz"`) {
		t.Errorf("missing Token; got %q", got)
	}
}

func TestDoJSON_401ReturnsErrUnauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		io.WriteString(w, "nope")
	}))
	defer srv.Close()

	c := NewClient(srv.URL, srv.Client())
	c.deviceID = "d"

	var out map[string]any
	err := c.doJSON(nil, "GET", "/x", nil, &out)
	if err == nil {
		t.Fatal("expected error")
	}
	if err != ErrUnauthorized {
		t.Errorf("want ErrUnauthorized, got %v", err)
	}
}

func TestDoJSON_5xxReturnsEmbyError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
		io.WriteString(w, "busy")
	}))
	defer srv.Close()

	c := NewClient(srv.URL, srv.Client())
	c.deviceID = "d"

	var out map[string]any
	err := c.doJSON(nil, "GET", "/x", nil, &out)
	var ee *EmbyError
	if err == nil {
		t.Fatal("expected error")
	}
	if !asEmbyError(err, &ee) || ee.Status != 503 {
		t.Errorf("want *EmbyError{Status:503}, got %v", err)
	}
}

func asEmbyError(err error, target **EmbyError) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*EmbyError); ok {
		*target = e
		return true
	}
	return false
}
