package downloader

import (
	"fmt"
	"path/filepath"
	"strings"
)

var illegal = []rune{'/', '\\', ':', '*', '?', '"', '<', '>', '|'}

func sanitize(s string) string {
	s = strings.TrimSpace(s)
	var b strings.Builder
	for _, r := range s {
		replace := false
		for _, ill := range illegal {
			if r == ill {
				replace = true
				break
			}
		}
		if replace {
			b.WriteRune('_')
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// MoviePaths returns (videoOutputPath, subtitleOutputPath) following the
// convention <output>/<moviesSubdir>/<Title> (<Year>)/<Title> (<Year>).<ext>
// and the subtitle placed next to it with a language code suffix.
func MoviePaths(outputDir, moviesSubdir, title string, year int, videoExt, lang, subExt string) (string, string) {
	safeTitle := sanitize(title)
	base := safeTitle
	if year > 0 {
		base = fmt.Sprintf("%s (%d)", safeTitle, year)
	}
	folder := filepath.Join(outputDir, moviesSubdir, base)
	video := filepath.Join(folder, base+"."+videoExt)
	if lang == "" {
		lang = "und"
	}
	sub := filepath.Join(folder, base+"."+lang+"."+subExt)
	return video, sub
}
