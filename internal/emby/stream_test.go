package emby

import (
	"strings"
	"testing"
)

func TestVideoDownloadURL(t *testing.T) {
	c := NewClient("https://emby.example.com", nil)
	c.SetSession(&Session{AccessToken: "tok-abc", UserID: "u"})
	got := c.VideoDownloadURL("item-1", "src-1")
	if !strings.HasPrefix(got, "https://emby.example.com/Items/item-1/Download") {
		t.Errorf("path: %q", got)
	}
	if !strings.Contains(got, "api_key=tok-abc") {
		t.Errorf("missing api_key: %q", got)
	}
	if !strings.Contains(got, "mediaSourceId=src-1") {
		t.Errorf("missing mediaSourceId: %q", got)
	}
}

func TestSubtitleDownloadURL(t *testing.T) {
	c := NewClient("https://emby.example.com", nil)
	c.SetSession(&Session{AccessToken: "tok-abc", UserID: "u"})
	got := c.SubtitleDownloadURL("item-1", "src-1", 2, "srt")
	want := "/Videos/item-1/src-1/Subtitles/2/Stream.srt"
	if !strings.Contains(got, want) {
		t.Errorf("\ngot  %s\nwant suffix %s", got, want)
	}
	if !strings.Contains(got, "api_key=tok-abc") {
		t.Errorf("missing api_key: %q", got)
	}
}
