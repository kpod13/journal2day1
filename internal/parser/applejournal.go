// Package parser provides Apple Journal HTML export parsing functionality.
package parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/html"

	"github.com/kpod13/journal2day1/internal/models"
)

// AppleJournalParser parses Apple Journal HTML exports.
type AppleJournalParser struct {
	basePath string
}

// NewAppleJournalParser creates a new parser for the given export directory.
func NewAppleJournalParser(basePath string) *AppleJournalParser {
	return &AppleJournalParser{basePath: basePath}
}

// ParseAll parses all entries from the Apple Journal export directory.
func (p *AppleJournalParser) ParseAll() ([]models.AppleJournalEntry, error) {
	entriesDir := filepath.Join(p.basePath, "Entries")

	files, err := os.ReadDir(entriesDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read entries directory")
	}

	entries := make([]models.AppleJournalEntry, 0, len(files))

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".html") {
			continue
		}

		entryPath := filepath.Join(entriesDir, file.Name())

		entry, err := p.ParseEntry(entryPath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse entry %s", file.Name())
		}

		entries = append(entries, *entry)
	}

	return entries, nil
}

// ParseEntry parses a single Apple Journal HTML entry.
func (p *AppleJournalParser) ParseEntry(filePath string) (*models.AppleJournalEntry, error) {
	file, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return nil, errors.Wrap(err, "failed to open file")
	}

	defer func() { _ = file.Close() }() //nolint:errcheck // read-only file close errors are not critical

	doc, err := html.Parse(file)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse HTML")
	}

	entry := &models.AppleJournalEntry{FilePath: filePath}
	p.extractFromNode(doc, entry)

	if entry.Date.IsZero() {
		entry.Date = p.extractDateFromAssets(entry.Assets)
	}

	if entry.Date.IsZero() {
		entry.Date = extractDateFromFilename(filePath)
	}

	return entry, nil
}

func (p *AppleJournalParser) extractDateFromAssets(assets []models.AppleJournalAsset) time.Time {
	if len(assets) == 0 {
		return time.Time{}
	}

	meta, err := p.LoadResourceMeta(assets[0].ID)
	if err != nil || meta.Date <= 0 {
		return time.Time{}
	}

	return models.CocoaTimestampToTime(meta.Date)
}

func (p *AppleJournalParser) extractFromNode(n *html.Node, entry *models.AppleJournalEntry) {
	if n.Type == html.ElementNode {
		p.processElement(n, entry)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		p.extractFromNode(c, entry)
	}
}

func (p *AppleJournalParser) processElement(n *html.Node, entry *models.AppleJournalEntry) {
	if n.Data == "div" {
		p.processDivElement(n, entry)
	}
}

func (p *AppleJournalParser) processDivElement(n *html.Node, entry *models.AppleJournalEntry) {
	class := getAttr(n, "class")

	switch {
	case strings.Contains(class, "pageHeader"):
		entry.Date = parsePageHeaderDate(getTextContent(n))
	case strings.Contains(class, "title"):
		entry.Title = strings.TrimSpace(getTextContent(n))
	case strings.Contains(class, "gridItem"):
		if asset := p.parseGridItem(n); asset != nil {
			entry.Assets = append(entry.Assets, *asset)
		}
	case strings.Contains(class, "bodyText"):
		if text := extractBodyText(n); text != "" {
			entry.Body = text
		}
	}
}

func parsePageHeaderDate(text string) time.Time {
	text = strings.TrimSpace(text)

	formats := []string{
		"Monday, 2 January 2006",
		"Monday, 02 January 2006",
		"2 January 2006",
		"02 January 2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, text); err == nil {
			return t
		}
	}

	return time.Time{}
}

func extractDateFromFilename(filePath string) time.Time {
	base := filepath.Base(filePath)
	re := regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})_`)

	matches := re.FindStringSubmatch(base)
	if len(matches) < 2 {
		return time.Time{}
	}

	t, err := time.Parse("2006-01-02", matches[1])
	if err != nil {
		return time.Time{}
	}

	return t
}

func (p *AppleJournalParser) parseGridItem(n *html.Node) *models.AppleJournalAsset {
	id := getAttr(n, "id")
	if id == "" {
		return nil
	}

	class := getAttr(n, "class")
	assetType := extractAssetType(class)

	filePath, ext := findMediaSrcInNode(n)
	if filePath == "" {
		filePath, ext = p.findResourceFile(id)
	}

	return &models.AppleJournalAsset{
		ID:        id,
		Type:      assetType,
		FilePath:  filePath,
		Extension: ext,
	}
}

func extractAssetType(class string) string {
	typeMap := map[string]string{
		"assetType_photo":          "photo",
		"assetType_livePhoto":      "photo",
		"assetType_video":          "video",
		"assetType_genericMap":     "map",
		"assetType_motionActivity": "activity",
		"assetType_audio":          "audio",
		"assetType_stateOfMind":    "stateOfMind",
	}

	for cssClass, assetType := range typeMap {
		if strings.Contains(class, cssClass) {
			return assetType
		}
	}

	return "unknown"
}

func findMediaSrcInNode(n *html.Node) (filePath, ext string) {
	if n.Type == html.ElementNode {
		filePath, ext = extractMediaSrc(n)
		if filePath != "" {
			return filePath, ext
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		filePath, ext = findMediaSrcInNode(c)
		if filePath != "" {
			return filePath, ext
		}
	}

	return "", ""
}

func extractMediaSrc(n *html.Node) (filePath, ext string) {
	switch n.Data {
	case "img":
		return extractImgSrc(n)
	case "video":
		return extractVideoSrc(n)
	}

	return "", ""
}

func extractImgSrc(n *html.Node) (filePath, ext string) {
	src := getAttr(n, "src")
	if src == "" || strings.Contains(src, "data:") {
		return "", ""
	}

	return src, strings.TrimPrefix(filepath.Ext(src), ".")
}

func extractVideoSrc(n *html.Node) (filePath, ext string) {
	src := getAttr(n, "src")
	if src == "" {
		src = findSourceElement(n)
	}

	if src == "" {
		return "", ""
	}

	return src, strings.TrimPrefix(filepath.Ext(src), ".")
}

func findSourceElement(n *html.Node) string {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "source" {
			return getAttr(c, "src")
		}
	}

	return ""
}

func (p *AppleJournalParser) findResourceFile(uuid string) (filePath, ext string) {
	resourcesDir := filepath.Join(p.basePath, "Resources")

	entries, err := os.ReadDir(resourcesDir)
	if err != nil {
		return "", ""
	}

	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), uuid) {
			continue
		}

		if strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		ext := strings.TrimPrefix(filepath.Ext(entry.Name()), ".")

		return filepath.Join("..", "Resources", entry.Name()), ext
	}

	return "", ""
}

func extractBodyText(n *html.Node) string {
	var parts []string

	collectBodyText(n, &parts)

	return strings.Join(parts, "\n")
}

func collectBodyText(n *html.Node, parts *[]string) {
	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" {
			*parts = append(*parts, text)
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collectBodyText(c, parts)
	}
}

// LoadResourceMeta loads the JSON metadata for a resource by UUID.
func (p *AppleJournalParser) LoadResourceMeta(uuid string) (*models.AppleJournalResourceMeta, error) {
	metaPath := filepath.Join(p.basePath, "Resources", uuid+".json")

	data, err := os.ReadFile(filepath.Clean(metaPath))
	if err != nil {
		return nil, errors.Wrap(err, "failed to read metadata")
	}

	var meta models.AppleJournalResourceMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, errors.Wrap(err, "failed to parse metadata")
	}

	return &meta, nil
}

// GetResourceFilePath returns the full path to a resource file.
func (p *AppleJournalParser) GetResourceFilePath(uuid string) string {
	resourcesDir := filepath.Join(p.basePath, "Resources")

	entries, err := os.ReadDir(resourcesDir)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), uuid) && !strings.HasSuffix(entry.Name(), ".json") {
			return filepath.Join(resourcesDir, entry.Name())
		}
	}

	return ""
}

func getAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}

	return ""
}

func getTextContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}

	var result strings.Builder

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result.WriteString(getTextContent(c))
	}

	return result.String()
}
