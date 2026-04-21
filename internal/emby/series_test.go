package emby

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestGetSeasons(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/Shows/series-99/Seasons") {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("UserId") != "user-1" {
			t.Errorf("UserId = %s", r.URL.Query().Get("UserId"))
		}
		data, _ := os.ReadFile("testdata/seasons.json")
		w.Write(data)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, srv.Client())
	c.SetSession(&Session{AccessToken: "tok", UserID: "user-1"})

	seasons, err := c.GetSeasons(context.Background(), "series-99")
	if err != nil {
		t.Fatal(err)
	}
	if len(seasons) != 3 {
		t.Fatalf("got %d seasons", len(seasons))
	}
	if seasons[0].IndexNumber != 1 || seasons[0].Type != TypeSeason {
		t.Errorf("seasons[0] = %+v", seasons[0])
	}
}
