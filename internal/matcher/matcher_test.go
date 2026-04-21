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

func TestPickSubtitles_ExternalOnlyByDefault(t *testing.T) {
	streams := []emby.MediaStream{
		{Index: 0, Type: "Video"},
		{Index: 1, Type: "Audio"},
		{Index: 2, Type: "Subtitle", Language: "zho", IsExternal: true},
		{Index: 3, Type: "Subtitle", Language: "zho", IsExternal: false}, // embedded — skipped
	}
	rule := SubtitleRule{Languages: []string{"zho"}, External: true}
	got := PickSubtitles(streams, rule)
	if len(got) != 1 {
		t.Fatalf("got %d, want 1 external zho", len(got))
	}
	if got[0].Index != 2 {
		t.Errorf("Index = %d, want 2", got[0].Index)
	}
}

func TestPickSubtitles_PreservesLanguageOrder(t *testing.T) {
	streams := []emby.MediaStream{
		{Index: 5, Type: "Subtitle", Language: "eng", IsExternal: true},
		{Index: 6, Type: "Subtitle", Language: "zho", IsExternal: true},
	}
	rule := SubtitleRule{Languages: []string{"zho", "eng"}, External: true}
	got := PickSubtitles(streams, rule)
	if len(got) != 2 {
		t.Fatalf("got %d", len(got))
	}
	if got[0].Language != "zho" || got[1].Language != "eng" {
		t.Errorf("order = %q, %q; want zho, eng", got[0].Language, got[1].Language)
	}
}

func TestPickSubtitles_SkipsUnmatchedLanguages(t *testing.T) {
	streams := []emby.MediaStream{
		{Index: 1, Type: "Subtitle", Language: "fra", IsExternal: true},
		{Index: 2, Type: "Subtitle", Language: "jpn", IsExternal: true},
	}
	rule := SubtitleRule{Languages: []string{"zho", "eng"}, External: true}
	got := PickSubtitles(streams, rule)
	if len(got) != 0 {
		t.Errorf("got %d, want 0", len(got))
	}
}

func TestPickSubtitles_MultipleStreamsSameLanguage(t *testing.T) {
	streams := []emby.MediaStream{
		{Index: 1, Type: "Subtitle", Language: "zho", IsExternal: true, Title: "Simplified"},
		{Index: 2, Type: "Subtitle", Language: "zho", IsExternal: true, Title: "Traditional"},
		{Index: 3, Type: "Subtitle", Language: "eng", IsExternal: true},
	}
	rule := SubtitleRule{Languages: []string{"zho"}, External: true}
	got := PickSubtitles(streams, rule)
	if len(got) != 2 {
		t.Fatalf("got %d, want 2 (both zho)", len(got))
	}
	if got[0].Index != 1 || got[1].Index != 2 {
		t.Errorf("indices = %d,%d", got[0].Index, got[1].Index)
	}
}

func TestPickSubtitles_EmptyInputReturnsEmpty(t *testing.T) {
	got := PickSubtitles(nil, SubtitleRule{Languages: []string{"zho"}, External: true})
	if len(got) != 0 {
		t.Errorf("got %d, want 0", len(got))
	}
}

func TestPickSubtitles_IncludeInternalWhenExternalFalse(t *testing.T) {
	streams := []emby.MediaStream{
		{Index: 1, Type: "Subtitle", Language: "zho", IsExternal: true},
		{Index: 2, Type: "Subtitle", Language: "zho", IsExternal: false},
	}
	rule := SubtitleRule{Languages: []string{"zho"}, External: false}
	got := PickSubtitles(streams, rule)
	if len(got) != 1 {
		t.Fatalf("got %d, want 1 internal", len(got))
	}
	if got[0].Index != 2 {
		t.Errorf("Index = %d, want 2", got[0].Index)
	}
}
