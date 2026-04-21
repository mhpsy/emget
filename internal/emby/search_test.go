package emby

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestSearch_MovieAndSeries(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/Users/user-1/Items") {
			t.Errorf("path = %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("SearchTerm") != "shaw" {
			t.Errorf("SearchTerm = %s", q.Get("SearchTerm"))
		}
		if q.Get("IncludeItemTypes") != "Movie,Series" {
			t.Errorf("IncludeItemTypes = %s", q.Get("IncludeItemTypes"))
		}
		if q.Get("Recursive") != "true" {
			t.Errorf("Recursive = %s", q.Get("Recursive"))
		}
		if q.Get("Limit") != "50" {
			t.Errorf("Limit = %s", q.Get("Limit"))
		}
		data, _ := os.ReadFile("testdata/search_movies.json")
		w.Write(data)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, srv.Client())
	c.SetSession(&Session{AccessToken: "tok", UserID: "user-1"})

	items, err := c.Search(context.Background(), "shaw", []ItemType{TypeMovie, TypeSeries}, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("want 2 items, got %d", len(items))
	}
	if items[0].Name != "The Shawshank Redemption" || items[0].Type != TypeMovie {
		t.Errorf("item[0] = %+v", items[0])
	}
}
