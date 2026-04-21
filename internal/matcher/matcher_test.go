package matcher

import (
	"errors"
	"testing"

	"github.com/mhpsy/emget/internal/emby"
)

func TestPickVersion_PrefersHigherResolution(t *testing.T) {
	sources := []emby.MediaSource{
		{ID: "a", Name: "720p WEB-DL", MediaStreams: []emby.MediaStream{{Type: "Video", Height: 720}}},
		{ID: "b", Name: "2160p BluRay", MediaStreams: []emby.MediaStream{{Type: "Video", Height: 2160}}},
		{ID: "c", Name: "1080p WEB-DL", MediaStreams: []emby.MediaStream{{Type: "Video", Height: 1080}}},
	}
	rule := VersionRule{ResolutionOrder: []int{2160, 1080, 720, 480}}
	got, err := PickVersion(sources, rule)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "b" {
		t.Errorf("got %q, want b", got.ID)
	}
}

func TestPickVersion_KeywordBreakstie(t *testing.T) {
	sources := []emby.MediaSource{
		{ID: "a", Name: "True Detective S01E01 1080p WEB-DL", MediaStreams: []emby.MediaStream{{Type: "Video", Height: 1080}}},
		{ID: "b", Name: "True Detective S01E01 1080p BluRay REMUX", MediaStreams: []emby.MediaStream{{Type: "Video", Height: 1080}}},
	}
	rule := VersionRule{
		ResolutionOrder: []int{1080},
		KeywordBoost:    []string{"BluRay", "REMUX"},
	}
	got, err := PickVersion(sources, rule)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "b" {
		t.Errorf("got %q, want b (two keyword hits beats zero)", got.ID)
	}
}

func TestPickVersion_FallsBackToSourceNameRegexForHeight(t *testing.T) {
	sources := []emby.MediaSource{
		{ID: "a", Name: "True Detective S01E01 1080p WEB-DL"},
		{ID: "b", Name: "True Detective S01E01 720p WEB-DL"},
	}
	rule := VersionRule{ResolutionOrder: []int{1080, 720}}
	got, err := PickVersion(sources, rule)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "a" {
		t.Errorf("got %q, want a", got.ID)
	}
}

func TestPickVersion_NoMatchReturnsError(t *testing.T) {
	sources := []emby.MediaSource{
		{ID: "a", Name: "mystery", MediaStreams: []emby.MediaStream{{Type: "Video", Height: 480}}},
	}
	rule := VersionRule{ResolutionOrder: []int{2160, 1080}}
	_, err := PickVersion(sources, rule)
	if !errors.Is(err, ErrNoMatch) {
		t.Fatalf("got %v, want ErrNoMatch", err)
	}
}

func TestPickVersion_TieBreaksByInputOrder(t *testing.T) {
	sources := []emby.MediaSource{
		{ID: "a", Name: "1080p BluRay", MediaStreams: []emby.MediaStream{{Type: "Video", Height: 1080}}},
		{ID: "b", Name: "1080p BluRay", MediaStreams: []emby.MediaStream{{Type: "Video", Height: 1080}}},
	}
	rule := VersionRule{ResolutionOrder: []int{1080}, KeywordBoost: []string{"BluRay"}}
	got, _ := PickVersion(sources, rule)
	if got.ID != "a" {
		t.Errorf("got %q, want a (stable sort, first wins)", got.ID)
	}
}

func TestPickVersion_EmptySources(t *testing.T) {
	_, err := PickVersion(nil, VersionRule{})
	if !errors.Is(err, ErrNoMatch) {
		t.Fatalf("got %v, want ErrNoMatch", err)
	}
}

func TestPickVersion_CaseInsensitiveKeywords(t *testing.T) {
	sources := []emby.MediaSource{
		{ID: "a", Name: "1080p BLURAY remux", MediaStreams: []emby.MediaStream{{Type: "Video", Height: 1080}}},
		{ID: "b", Name: "1080p web-dl", MediaStreams: []emby.MediaStream{{Type: "Video", Height: 1080}}},
	}
	rule := VersionRule{ResolutionOrder: []int{1080}, KeywordBoost: []string{"BluRay", "REMUX"}}
	got, _ := PickVersion(sources, rule)
	if got.ID != "a" {
		t.Errorf("got %q, want a (case insensitive match)", got.ID)
	}
}
