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

func TestGetEpisodes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/Shows/series-99/Episodes") {
			t.Errorf("path = %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("SeasonId") != "season-1" {
			t.Errorf("SeasonId = %s", q.Get("SeasonId"))
		}
		if q.Get("UserId") != "user-1" {
			t.Errorf("UserId = %s", q.Get("UserId"))
		}
		if !strings.Contains(q.Get("Fields"), "MediaSources") {
			t.Errorf("Fields missing MediaSources: %s", q.Get("Fields"))
		}
		data, _ := os.ReadFile("testdata/episodes_season1.json")
		w.Write(data)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, srv.Client())
	c.SetSession(&Session{AccessToken: "tok", UserID: "user-1"})

	episodes, err := c.GetEpisodes(context.Background(), "series-99", "season-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(episodes) != 2 {
		t.Fatalf("got %d episodes", len(episodes))
	}
	if episodes[0].IndexNumber != 1 || episodes[0].ParentIndexNumber != 1 {
		t.Errorf("episodes[0] index = %d, parent = %d", episodes[0].IndexNumber, episodes[0].ParentIndexNumber)
	}
	if len(episodes[0].MediaSources) != 2 {
		t.Errorf("episodes[0] sources = %d", len(episodes[0].MediaSources))
	}
	if len(episodes[0].MediaSources[0].MediaStreams) != 4 {
		t.Errorf("streams = %d", len(episodes[0].MediaSources[0].MediaStreams))
	}
}
