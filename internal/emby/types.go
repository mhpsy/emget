package emby

import "time"

type ItemType string

const (
	TypeMovie   ItemType = "Movie"
	TypeSeries  ItemType = "Series"
	TypeSeason  ItemType = "Season"
	TypeEpisode ItemType = "Episode"
)

type AuthResult struct {
	User        AuthUser `json:"User"`
	AccessToken string   `json:"AccessToken"`
	ServerID    string   `json:"ServerId"`
}

type AuthUser struct {
	ID   string `json:"Id"`
	Name string `json:"Name"`
}

type SearchResponse struct {
	Items            []Item `json:"Items"`
	TotalRecordCount int    `json:"TotalRecordCount"`
}

type Item struct {
	ID               string         `json:"Id"`
	Name             string         `json:"Name"`
	Type             ItemType       `json:"Type"`
	ProductionYear   int            `json:"ProductionYear,omitempty"`
	ParentIndexNumber int           `json:"ParentIndexNumber,omitempty"` // Season #
	IndexNumber      int            `json:"IndexNumber,omitempty"`       // Episode #
	SeriesName       string         `json:"SeriesName,omitempty"`
	SeriesID         string         `json:"SeriesId,omitempty"`
	RunTimeTicks     int64          `json:"RunTimeTicks,omitempty"`
	Overview         string         `json:"Overview,omitempty"`
	MediaSources     []MediaSource  `json:"MediaSources,omitempty"`
	DateCreated      time.Time      `json:"DateCreated,omitempty"`
}

type MediaSource struct {
	ID           string        `json:"Id"`
	Name         string        `json:"Name"`
	Path         string        `json:"Path,omitempty"`
	Container    string        `json:"Container,omitempty"`
	Size         int64         `json:"Size,omitempty"`
	MediaStreams []MediaStream `json:"MediaStreams,omitempty"`
}

type MediaStream struct {
	Index       int    `json:"Index"`
	Codec       string `json:"Codec"`
	Language    string `json:"Language,omitempty"`
	Type        string `json:"Type"` // Video | Audio | Subtitle
	Title       string `json:"Title,omitempty"`
	DisplayTitle string `json:"DisplayTitle,omitempty"`
	Height      int    `json:"Height,omitempty"`
	Width       int    `json:"Width,omitempty"`
	IsDefault   bool   `json:"IsDefault,omitempty"`
	IsForced    bool   `json:"IsForced,omitempty"`
	IsExternal  bool   `json:"IsExternal,omitempty"`
	CodecTag    string `json:"CodecTag,omitempty"`
	// For external subtitles:
	DeliveryURL string `json:"DeliveryUrl,omitempty"`
}
