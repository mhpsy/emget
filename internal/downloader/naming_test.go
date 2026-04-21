package downloader

import "testing"

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
	if out != wantVideo {
		t.Errorf("video\n got %q\nwant %q", out, wantVideo)
	}
	if sub != wantSub {
		t.Errorf("sub\n got %q\nwant %q", sub, wantSub)
	}
}

func TestMoviePaths_UnknownYear(t *testing.T) {
	out, _ := MoviePaths("/media", "Movies", "Some Film", 0, "mp4", "eng", "ass")
	want := "/media/Movies/Some Film/Some Film.mp4"
	if out != want {
		t.Errorf("\n got %q\nwant %q", out, want)
	}
}

func TestMoviePaths_UnknownLanguage(t *testing.T) {
	_, sub := MoviePaths("/media", "Movies", "X", 2020, "mkv", "", "srt")
	want := "/media/Movies/X (2020)/X (2020).und.srt"
	if sub != want {
		t.Errorf("\n got %q\nwant %q", sub, want)
	}
}

func TestMoviePaths_SanitizesTitle(t *testing.T) {
	out, _ := MoviePaths("/media", "Movies", "Bad:Title?", 2020, "mkv", "eng", "srt")
	want := "/media/Movies/Bad_Title_ (2020)/Bad_Title_ (2020).mkv"
	if out != want {
		t.Errorf("\n got %q\nwant %q", out, want)
	}
}
