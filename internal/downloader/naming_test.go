package downloader

import (
	"path/filepath"
	"testing"
)

// toSlash normalizes native-separator paths (backslashes on Windows) to
// forward slashes so tests can compare against slash-form literals on
// all platforms. filepath.Join correctly produces \ on Windows; we
// translate only at the assertion boundary.
func toSlash(p string) string { return filepath.ToSlash(p) }

func TestSanitize(t *testing.T) {
	cases := []struct{ in, want string }{
		{"A/B:C?D", "A_B_C_D"},
		{"  trim  ", "trim"},
		{"normal name", "normal name"},
		{`a"b|c<d>e`, "a_b_c_d_e"},
	}
	for _, tc := range cases {
		if got := sanitize(tc.in); got != tc.want {
			t.Errorf("sanitize(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestMoviePaths(t *testing.T) {
	out, sub := MoviePaths("/media", "Movies", "The Shawshank Redemption", 1994, "mkv", "zho", "srt")
	wantVideo := "/media/Movies/The Shawshank Redemption (1994)/The Shawshank Redemption (1994).mkv"
	wantSub := "/media/Movies/The Shawshank Redemption (1994)/The Shawshank Redemption (1994).zho.srt"
	if toSlash(out) != wantVideo {
		t.Errorf("video\n got %q\nwant %q", out, wantVideo)
	}
	if toSlash(sub) != wantSub {
		t.Errorf("sub\n got %q\nwant %q", sub, wantSub)
	}
}

func TestMoviePaths_UnknownYear(t *testing.T) {
	out, _ := MoviePaths("/media", "Movies", "Some Film", 0, "mp4", "eng", "ass")
	want := "/media/Movies/Some Film/Some Film.mp4"
	if toSlash(out) != want {
		t.Errorf("\n got %q\nwant %q", out, want)
	}
}

func TestMoviePaths_UnknownLanguage(t *testing.T) {
	_, sub := MoviePaths("/media", "Movies", "X", 2020, "mkv", "", "srt")
	want := "/media/Movies/X (2020)/X (2020).und.srt"
	if toSlash(sub) != want {
		t.Errorf("\n got %q\nwant %q", sub, want)
	}
}

func TestMoviePaths_SanitizesTitle(t *testing.T) {
	out, _ := MoviePaths("/media", "Movies", "Bad:Title?", 2020, "mkv", "eng", "srt")
	want := "/media/Movies/Bad_Title_ (2020)/Bad_Title_ (2020).mkv"
	if toSlash(out) != want {
		t.Errorf("\n got %q\nwant %q", out, want)
	}
}

func TestTVPaths(t *testing.T) {
	video, sub := TVPaths("/media", "TV", "True Detective", 1, 3, "The Locked Room", "mkv", "zho", "srt")
	wantVideo := "/media/TV/True Detective/Season 01/True Detective - S01E03 - The Locked Room.mkv"
	wantSub := "/media/TV/True Detective/Season 01/True Detective - S01E03 - The Locked Room.zho.srt"
	if toSlash(video) != wantVideo {
		t.Errorf("video\n got %q\nwant %q", video, wantVideo)
	}
	if toSlash(sub) != wantSub {
		t.Errorf("sub\n got %q\nwant %q", sub, wantSub)
	}
}

func TestTVPaths_SanitizesSeriesAndTitle(t *testing.T) {
	video, _ := TVPaths("/media", "TV", "Bad:Series?", 2, 10, "Ep/Title", "mkv", "eng", "srt")
	want := "/media/TV/Bad_Series_/Season 02/Bad_Series_ - S02E10 - Ep_Title.mkv"
	if toSlash(video) != want {
		t.Errorf("\n got %q\nwant %q", video, want)
	}
}

func TestTVPaths_UnknownLanguage(t *testing.T) {
	_, sub := TVPaths("/media", "TV", "X", 1, 1, "Pilot", "mkv", "", "srt")
	want := "/media/TV/X/Season 01/X - S01E01 - Pilot.und.srt"
	if toSlash(sub) != want {
		t.Errorf("\n got %q\nwant %q", sub, want)
	}
}

func TestTVPaths_EmptyEpisodeTitle(t *testing.T) {
	video, _ := TVPaths("/media", "TV", "Show", 1, 2, "", "mkv", "zho", "srt")
	want := "/media/TV/Show/Season 01/Show - S01E02.mkv"
	if toSlash(video) != want {
		t.Errorf("\n got %q\nwant %q", video, want)
	}
}

func TestTVPaths_HighSeasonEpisodeNumbers(t *testing.T) {
	video, _ := TVPaths("/media", "TV", "Big Show", 12, 100, "Finale", "mkv", "eng", "srt")
	want := "/media/TV/Big Show/Season 12/Big Show - S12E100 - Finale.mkv"
	if toSlash(video) != want {
		t.Errorf("\n got %q\nwant %q", video, want)
	}
}
