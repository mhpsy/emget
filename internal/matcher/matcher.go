// Package matcher provides pure-function rule engines for selecting among
// Emby media sources and media streams. Inputs come from emby DTOs; outputs
// are pointers into those same slices (no copying).
package matcher

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/mhpsy/emget/internal/emby"
)

// ErrNoMatch is returned when no candidate satisfies the rule.
var ErrNoMatch = errors.New("matcher: no match")

// VersionRule describes how to pick a MediaSource from a list of candidates.
type VersionRule struct {
	ResolutionOrder []int
	KeywordBoost    []string
}

var heightRegex = regexp.MustCompile(`\b(\d{3,4})p\b`)

// PickVersion returns the highest-scored source according to rule. Sources
// with score 0 (no matching resolution AND no matching keyword) are rejected;
// if all sources score 0, returns ErrNoMatch. Ties are broken by input order.
func PickVersion(sources []emby.MediaSource, rule VersionRule) (*emby.MediaSource, error) {
	bestIdx := -1
	bestScore := 0
	for i := range sources {
		score := scoreVersion(&sources[i], rule)
		if score > bestScore {
			bestScore = score
			bestIdx = i
		}
	}
	if bestIdx < 0 {
		return nil, ErrNoMatch
	}
	return &sources[bestIdx], nil
}

func scoreVersion(src *emby.MediaSource, rule VersionRule) int {
	height := sourceHeight(src)
	resScore := 0
	if height > 0 {
		for i, target := range rule.ResolutionOrder {
			if height == target {
				resScore = len(rule.ResolutionOrder) - i
				break
			}
		}
	}
	kwScore := 0
	lowerName := strings.ToLower(src.Name)
	for _, kw := range rule.KeywordBoost {
		if strings.Contains(lowerName, strings.ToLower(kw)) {
			kwScore++
		}
	}
	return resScore*1000 + kwScore
}

func sourceHeight(src *emby.MediaSource) int {
	for _, s := range src.MediaStreams {
		if s.Type == "Video" && s.Height > 0 {
			return s.Height
		}
	}
	if m := heightRegex.FindStringSubmatch(src.Name); len(m) == 2 {
		if v, err := strconv.Atoi(m[1]); err == nil {
			return v
		}
	}
	return 0
}

// SubtitleRule filters MediaStreams of type "Subtitle".
type SubtitleRule struct {
	Languages []string
	External  bool
}

// PickSubtitles returns MediaStreams of type "Subtitle" where IsExternal==rule.External
// and Language is in rule.Languages. Output is stable-sorted by the position of
// the language in rule.Languages (primary) then by the stream's original order
// (secondary). Non-matching streams are dropped.
func PickSubtitles(streams []emby.MediaStream, rule SubtitleRule) []emby.MediaStream {
	langOrder := map[string]int{}
	for i, l := range rule.Languages {
		langOrder[l] = i
	}
	type indexed struct {
		pos    int
		stream emby.MediaStream
	}
	var matched []indexed
	for i, s := range streams {
		if s.Type != "Subtitle" {
			continue
		}
		if s.IsExternal != rule.External {
			continue
		}
		if _, ok := langOrder[s.Language]; !ok {
			continue
		}
		matched = append(matched, indexed{pos: i, stream: s})
	}
	// Sort by language order, then by original position (stable).
	for i := 1; i < len(matched); i++ {
		for j := i; j > 0; j-- {
			a, b := matched[j-1], matched[j]
			ai, bi := langOrder[a.stream.Language], langOrder[b.stream.Language]
			if ai < bi {
				break
			}
			if ai == bi && a.pos <= b.pos {
				break
			}
			matched[j-1], matched[j] = b, a
		}
	}
	out := make([]emby.MediaStream, len(matched))
	for i, m := range matched {
		out[i] = m.stream
	}
	return out
}
