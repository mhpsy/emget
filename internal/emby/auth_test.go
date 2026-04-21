package emby

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestAuthenticate_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Users/AuthenticateByName" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("method = %s", r.Method)
		}
		auth := r.Header.Get("X-Emby-Authorization")
		if !strings.Contains(auth, `Client="emget"`) {
			t.Errorf("bad auth header: %s", auth)
		}
		if strings.Contains(auth, "Token=") {
			t.Errorf("pre-auth should not carry Token: %s", auth)
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["Username"] != "mhpsy" || body["Pw"] != "secret" {
			t.Errorf("bad body: %v", body)
		}
		data, _ := os.ReadFile("testdata/auth_ok.json")
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, srv.Client())
	c.SetDeviceID("dev-1")

	sess, err := c.Authenticate(context.Background(), "mhpsy", "secret")
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if sess.AccessToken != "tok-abc" {
		t.Errorf("token = %q", sess.AccessToken)
	}
	if sess.UserID != "user-123" {
		t.Errorf("user_id = %q", sess.UserID)
	}
	if sess.ServerID != "srv-xyz" {
		t.Errorf("server_id = %q", sess.ServerID)
	}
	if sess.DeviceID != "dev-1" {
		t.Errorf("device_id = %q", sess.DeviceID)
	}
}

func TestAuthenticate_BadPassword(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, srv.Client())
	c.SetDeviceID("dev-1")

	_, err := c.Authenticate(context.Background(), "u", "wrong")
	if err != ErrUnauthorized {
		t.Fatalf("want ErrUnauthorized, got %v", err)
	}
}
