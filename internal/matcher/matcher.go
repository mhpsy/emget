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
