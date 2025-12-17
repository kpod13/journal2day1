// Package converter provides conversion from Apple Journal to DayOne format.
package converter

import (
	"archive/zip"
	"crypto/md5" //nolint:gosec // MD5 is required by DayOne format specification
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/kpod13/journal2day1/internal/models"
	"github.com/kpod13/journal2day1/internal/parser"
)

const (
	iso8601Format  = "2006-01-02T15:04:05Z"
	dayOneVersion  = "1.0"
	dirPermission  = 0o750
	filePermission = 0o600
)

// ProgressFunc is called during conversion to report progress.
type ProgressFunc func(current, total int)

// Converter converts Apple Journal entries to DayOne format.
type Converter struct {
	parser      *parser.AppleJournalParser
	journalName string
	timeZone    string
	onProgress  ProgressFunc
}

// NewConverter creates a new converter.
func NewConverter(appleJournalPath, journalName string) *Converter {
	return &Converter{
		parser:      parser.NewAppleJournalParser(appleJournalPath),
		journalName: journalName,
		timeZone:    "Europe/Sofia",
	}
}

// SetTimeZone sets the timezone for entries.
func (c *Converter) SetTimeZone(tz string) {
	c.timeZone = tz
}

// SetProgressFunc sets the progress callback function.
func (c *Converter) SetProgressFunc(fn ProgressFunc) {
	c.onProgress = fn
}

// Convert converts all Apple Journal entries and creates a DayOne ZIP archive.
func (c *Converter) Convert(outputPath string) error {
	entries, err := c.parser.ParseAll()
	if err != nil {
		return errors.Wrap(err, "failed to parse entries")
	}

	tmpDir, err := os.MkdirTemp("", "journal2day1-*")
	if err != nil {
		return errors.Wrap(err, "failed to create temp dir")
	}
	defer os.RemoveAll(tmpDir)

	dirs, err := c.createOutputDirs(tmpDir)
	if err != nil {
		return err
	}

	dayOneExport := c.convertEntries(entries, dirs)

	if err := c.writeJSON(tmpDir, dayOneExport); err != nil {
		return err
	}

	return createZipArchive(tmpDir, outputPath)
}

type outputDirs struct {
	photos string
	videos string
}

func (c *Converter) createOutputDirs(tmpDir string) (*outputDirs, error) {
	dirs := &outputDirs{
		photos: filepath.Join(tmpDir, "photos"),
		videos: filepath.Join(tmpDir, "videos"),
	}

	if err := os.MkdirAll(dirs.photos, dirPermission); err != nil {
		return nil, errors.Wrap(err, "failed to create photos dir")
	}

	if err := os.MkdirAll(dirs.videos, dirPermission); err != nil {
		return nil, errors.Wrap(err, "failed to create videos dir")
	}

	return dirs, nil
}

func (c *Converter) convertEntries(entries []models.AppleJournalEntry, dirs *outputDirs) models.DayOneExport {
	dayOneExport := models.DayOneExport{
		Metadata: models.DayOneMetadata{Version: dayOneVersion},
		Entries:  make([]models.DayOneEntry, 0, len(entries)),
	}

	total := len(entries)

	for i := range entries {
		if c.onProgress != nil {
			c.onProgress(i+1, total)
		}

		dayOneEntry := c.convertEntry(&entries[i], dirs)
		dayOneExport.Entries = append(dayOneExport.Entries, *dayOneEntry)
	}

	return dayOneExport
}

func (c *Converter) writeJSON(tmpDir string, export models.DayOneExport) error {
	jsonPath := filepath.Join(tmpDir, c.journalName+".json")

	jsonData, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal JSON")
	}

	if err := os.WriteFile(jsonPath, jsonData, filePermission); err != nil {
		return errors.Wrap(err, "failed to write JSON")
	}

	return nil
}

func (c *Converter) convertEntry(entry *models.AppleJournalEntry, dirs *outputDirs) *models.DayOneEntry {
	now := time.Now().UTC().Format(iso8601Format)
	creationDate := entry.Date.UTC().Format(iso8601Format)

	dayOneEntry := &models.DayOneEntry{
		UUID:           strings.ToUpper(strings.ReplaceAll(uuid.NewString(), "-", "")),
		CreationDate:   creationDate,
		ModifiedDate:   now,
		Starred:        false,
		IsPinned:       false,
		IsAllDay:       false,
		Duration:       0,
		TimeZone:       c.timeZone,
		CreationDevice: "journal2day1",
	}

	photos, videos, photoRefs := c.processAssets(entry, dirs, creationDate)
	dayOneEntry.Photos = photos
	dayOneEntry.Videos = videos
	dayOneEntry.Text = buildEntryText(entry, photoRefs)

	return dayOneEntry
}

func (c *Converter) processAssets(
	entry *models.AppleJournalEntry,
	dirs *outputDirs,
	creationDate string,
) ([]models.DayOnePhoto, []models.DayOneVideo, []string) {
	var (
		photos    []models.DayOnePhoto
		videos    []models.DayOneVideo
		photoRefs []string
	)

	for i, asset := range entry.Assets {
		if shouldSkipAsset(asset.Type) {
			continue
		}

		photo, video, ref := c.processAsset(asset, i, dirs, creationDate)
		if photo != nil {
			photos = append(photos, *photo)
			photoRefs = append(photoRefs, ref)
		}

		if video != nil {
			videos = append(videos, *video)
		}
	}

	return photos, videos, photoRefs
}

func shouldSkipAsset(assetType string) bool {
	skipTypes := map[string]bool{
		"map":         true,
		"activity":    true,
		"stateOfMind": true,
	}

	return skipTypes[assetType]
}

func (c *Converter) processAsset(
	asset models.AppleJournalAsset,
	order int,
	dirs *outputDirs,
	creationDate string,
) (*models.DayOnePhoto, *models.DayOneVideo, string) {
	resourcePath := c.parser.GetResourceFilePath(asset.ID)
	if resourcePath == "" {
		return nil, nil, ""
	}

	md5Hash, fileSize, err := copyMediaFile(resourcePath, asset.Extension, dirs)
	if err != nil {
		return nil, nil, ""
	}

	assetDate := c.getAssetDate(asset.ID, creationDate)
	identifier := strings.ToUpper(strings.ReplaceAll(asset.ID, "-", ""))
	ext := strings.ToLower(asset.Extension)

	if isVideoExtension(ext) {
		video := createVideo(identifier, ext, md5Hash, fileSize, order, assetDate)
		return nil, video, ""
	}

	photo := createPhoto(identifier, ext, md5Hash, fileSize, order, assetDate)
	ref := fmt.Sprintf("![](dayone-moment://%s)", identifier)

	return photo, nil, ref
}

func (c *Converter) getAssetDate(assetID, fallbackDate string) string {
	meta, err := c.parser.LoadResourceMeta(assetID)
	if err != nil || meta.Date <= 0 {
		return fallbackDate
	}

	return models.CocoaTimestampToTime(meta.Date).UTC().Format(iso8601Format)
}

func createPhoto(id, ext, md5Hash string, size int64, order int, date string) *models.DayOnePhoto {
	return &models.DayOnePhoto{
		Identifier:     id,
		Type:           normalizeExtension(ext),
		MD5:            md5Hash,
		FileSize:       size,
		OrderInEntry:   order,
		CreationDevice: "journal2day1",
		Duration:       0,
		Favorite:       false,
		IsSketch:       false,
		Date:           date,
	}
}

func createVideo(id, ext, md5Hash string, size int64, order int, date string) *models.DayOneVideo {
	return &models.DayOneVideo{
		Identifier:     id,
		Type:           normalizeExtension(ext),
		MD5:            md5Hash,
		FileSize:       size,
		OrderInEntry:   order,
		CreationDevice: "journal2day1",
		Duration:       0,
		Favorite:       false,
		Date:           date,
	}
}

func buildEntryText(entry *models.AppleJournalEntry, photoRefs []string) string {
	var textParts []string

	if entry.Title != "" {
		textParts = append(textParts, "# "+entry.Title)
	}

	if entry.Body != "" {
		textParts = append(textParts, entry.Body)
	}

	if len(photoRefs) > 0 {
		textParts = append(textParts, strings.Join(photoRefs, "\n"))
	}

	return strings.Join(textParts, "\n\n")
}

func copyMediaFile(srcPath, ext string, dirs *outputDirs) (md5Hash string, fileSize int64, err error) {
	src, err := os.Open(srcPath)
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to open source")
	}
	defer src.Close()

	md5Hash, err = calculateMD5(src)
	if err != nil {
		return "", 0, err
	}

	stat, err := src.Stat()
	if err != nil {
		return "", 0, errors.Wrap(err, "failed to get file stat")
	}

	if _, err := src.Seek(0, 0); err != nil {
		return "", 0, errors.Wrap(err, "failed to seek file")
	}

	dstPath := getDestinationPath(ext, md5Hash, dirs)

	if err := copyToFile(src, dstPath); err != nil {
		return "", 0, err
	}

	return md5Hash, stat.Size(), nil
}

func calculateMD5(r io.Reader) (string, error) {
	hash := md5.New() //nolint:gosec // MD5 is required by DayOne format specification

	if _, err := io.Copy(hash, r); err != nil {
		return "", errors.Wrap(err, "failed to calculate MD5")
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func getDestinationPath(ext, md5Hash string, dirs *outputDirs) string {
	normalizedExt := normalizeExtension(strings.ToLower(ext))

	if isVideoExtension(ext) {
		return filepath.Join(dirs.videos, md5Hash+"."+normalizedExt)
	}

	return filepath.Join(dirs.photos, md5Hash+"."+normalizedExt)
}

func copyToFile(src io.Reader, dstPath string) error {
	dst, err := os.Create(dstPath)
	if err != nil {
		return errors.Wrap(err, "failed to create destination")
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return errors.Wrap(err, "failed to copy file")
	}

	return nil
}

func createZipArchive(srcDir, dstPath string) error {
	zipFile, err := os.Create(dstPath)
	if err != nil {
		return errors.Wrap(err, "failed to create ZIP file")
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		return addFileToZip(archive, srcDir, path, info)
	})
}

func addFileToZip(archive *zip.Writer, srcDir, path string, info os.FileInfo) error {
	relPath, err := filepath.Rel(srcDir, path)
	if err != nil {
		return errors.Wrap(err, "failed to get relative path")
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return errors.Wrap(err, "failed to create ZIP header")
	}

	header.Name = relPath
	header.Method = zip.Deflate

	writer, err := archive.CreateHeader(header)
	if err != nil {
		return errors.Wrap(err, "failed to create ZIP entry")
	}

	file, err := os.Open(path)
	if err != nil {
		return errors.Wrap(err, "failed to open file for ZIP")
	}
	defer file.Close()

	if _, err := io.Copy(writer, file); err != nil {
		return errors.Wrap(err, "failed to write to ZIP")
	}

	return nil
}

func isVideoExtension(ext string) bool {
	videoExts := map[string]bool{
		"mov": true,
		"mp4": true,
		"m4v": true,
		"avi": true,
	}

	return videoExts[strings.ToLower(ext)]
}

func normalizeExtension(ext string) string {
	if strings.EqualFold(ext, "jpg") {
		return "jpeg"
	}

	return strings.ToLower(ext)
}
