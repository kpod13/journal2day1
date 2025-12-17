// Package models provides data structures for Apple Journal and DayOne formats.
package models

import "time"

// AppleJournalEntry represents a parsed entry from Apple Journal HTML export.
type AppleJournalEntry struct {
	Date     time.Time
	Title    string
	Body     string
	Assets   []AppleJournalAsset
	FilePath string
}

// AppleJournalAsset represents a media asset in an Apple Journal entry.
type AppleJournalAsset struct {
	ID        string
	Type      string
	FilePath  string
	Extension string
}

// AppleJournalResourceMeta represents the JSON metadata for a resource.
type AppleJournalResourceMeta struct {
	Date      float64 `json:"date"`
	PlaceName string  `json:"placeName"`
}

// appleCocoaEpoch is the reference date for Apple/Cocoa timestamps (2001-01-01).
var appleCocoaEpoch = time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)

// CocoaTimestampToTime converts Apple/Cocoa timestamp to time.Time.
func CocoaTimestampToTime(timestamp float64) time.Time {
	return appleCocoaEpoch.Add(time.Duration(timestamp * float64(time.Second)))
}
