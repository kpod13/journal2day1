package models

// DayOneExport represents the root structure of a DayOne JSON export.
type DayOneExport struct {
	Metadata DayOneMetadata `json:"metadata"`
	Entries  []DayOneEntry  `json:"entries"`
}

// DayOneMetadata contains export metadata.
type DayOneMetadata struct {
	Version string `json:"version"`
}

// DayOneEntry represents a single journal entry in DayOne format.
type DayOneEntry struct {
	UUID           string          `json:"uuid"`
	CreationDate   string          `json:"creationDate"` // ISO 8601 format
	ModifiedDate   string          `json:"modifiedDate"` // ISO 8601 format
	Text           string          `json:"text"`         // Markdown content
	RichText       string          `json:"richText,omitempty"`
	Starred        bool            `json:"starred"`
	IsPinned       bool            `json:"isPinned"`
	IsAllDay       bool            `json:"isAllDay"`
	Duration       int             `json:"duration"`
	TimeZone       string          `json:"timeZone"`
	CreationDevice string          `json:"creationDevice,omitempty"`
	Photos         []DayOnePhoto   `json:"photos,omitempty"`
	Videos         []DayOneVideo   `json:"videos,omitempty"`
	Location       *DayOneLocation `json:"location,omitempty"`
}

// DayOnePhoto represents a photo attachment in DayOne.
type DayOnePhoto struct {
	Identifier     string               `json:"identifier"` // UUID without dashes, uppercase
	Type           string               `json:"type"`       // jpeg, heic, png, etc.
	MD5            string               `json:"md5"`
	FileSize       int64                `json:"fileSize"`
	OrderInEntry   int                  `json:"orderInEntry"`
	CreationDevice string               `json:"creationDevice,omitempty"`
	Duration       int                  `json:"duration"`
	Favorite       bool                 `json:"favorite"`
	IsSketch       bool                 `json:"isSketch"`
	Date           string               `json:"date"` // ISO 8601 format
	Width          int                  `json:"width,omitempty"`
	Height         int                  `json:"height,omitempty"`
	Location       *DayOnePhotoLocation `json:"location,omitempty"`
}

// DayOneVideo represents a video attachment in DayOne.
type DayOneVideo struct {
	Identifier     string `json:"identifier"` // UUID without dashes, uppercase
	Type           string `json:"type"`       // mov, mp4, etc.
	MD5            string `json:"md5"`
	FileSize       int64  `json:"fileSize"`
	OrderInEntry   int    `json:"orderInEntry"`
	CreationDevice string `json:"creationDevice,omitempty"`
	Duration       int    `json:"duration"`
	Favorite       bool   `json:"favorite"`
	Date           string `json:"date"` // ISO 8601 format
	Width          int    `json:"width,omitempty"`
	Height         int    `json:"height,omitempty"`
}

// DayOneLocation represents location information for an entry.
type DayOneLocation struct {
	PlaceName          string        `json:"placeName,omitempty"`
	LocalityName       string        `json:"localityName,omitempty"`
	Country            string        `json:"country,omitempty"`
	AdministrativeArea string        `json:"administrativeArea,omitempty"`
	Longitude          float64       `json:"longitude,omitempty"`
	Latitude           float64       `json:"latitude,omitempty"`
	Region             *DayOneRegion `json:"region,omitempty"`
}

// DayOnePhotoLocation represents location information for a photo.
type DayOnePhotoLocation struct {
	TimeZoneName string  `json:"timeZoneName,omitempty"`
	Longitude    float64 `json:"longitude,omitempty"`
	Latitude     float64 `json:"latitude,omitempty"`
}

// DayOneRegion represents a geographic region.
type DayOneRegion struct {
	Center DayOneCenter `json:"center"`
	Radius float64      `json:"radius"`
}

// DayOneCenter represents the center of a geographic region.
type DayOneCenter struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}
