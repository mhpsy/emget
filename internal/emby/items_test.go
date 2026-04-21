package emby

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestGetItem_MovieWithMediaSources(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/Users/user-1/Items/m-1") {
			t.Errorf("path = %s", r.URL.Path)
		}
		if !strings.Contains(r.URL.RawQuery, "MediaSources") {
			t.Errorf("missing Fields=...MediaSources... in %s", r.URL.RawQuery)
		}
		data, _ := os.ReadFile("testdata/item_movie.json")
		w.Write(data)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, srv.Client())
	c.SetSession(&Session{AccessToken: "tok", UserID: "user-1"})

	item, err := c.GetItem(context.Background(), "m-1")
	if err != nil {
		t.Fatal(err)
	}
	if item.Name != "The Shawshank Redemption" {
		t.Errorf("name = %q", item.Name)
	}
	if len(item.MediaSources) != 2 {
		t.Fatalf("media sources = %d", len(item.MediaSources))
	}
	if item.MediaSources[0].ID != "src-1" {
		t.Errorf("first source id = %q", item.MediaSources[0].ID)
	}
	if len(item.MediaSources[0].MediaStreams) != 4 {
		t.Errorf("streams = %d", len(item.MediaSources[0].MediaStreams))
	}
}
