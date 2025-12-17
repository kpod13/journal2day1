package converter_test

import (
	"archive/zip"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kpod13/journal2day1/internal/converter"
	"github.com/kpod13/journal2day1/internal/models"
)

func TestConvert(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	outputPath := filepath.Join(tmpDir, "output.zip")

	setupConvertTestData(t, inputDir)

	conv := converter.NewConverter(inputDir, "TestJournal")
	conv.SetTimeZone("Europe/Sofia")

	err := conv.Convert(outputPath)
	require.NoError(t, err)

	require.FileExists(t, outputPath)

	verifyZipContents(t, outputPath)
}

func TestSetTimeZone(t *testing.T) {
	t.Parallel()

	conv := converter.NewConverter("/fake/path", "Test")
	conv.SetTimeZone("America/New_York")

	// Just verify it doesn't panic - actual timezone usage is tested in Convert
	require.NotNil(t, conv)
}

func setupConvertTestData(t *testing.T, inputDir string) {
	t.Helper()

	entriesDir := filepath.Join(inputDir, "Entries")
	resourcesDir := filepath.Join(inputDir, "Resources")

	require.NoError(t, os.MkdirAll(entriesDir, 0o750))
	require.NoError(t, os.MkdirAll(resourcesDir, 0o750))

	htmlContent := `<!DOCTYPE html>
<html>
<body>
<div class="pageHeader">Monday, 15 December 2025</div>
<div class="assetGrid">
    <div id="TEST-UUID-1234-5678-ABCD" class="gridItem assetType_photo">
        <img src="../Resources/TEST-UUID-1234-5678-ABCD.jpg" class="asset_image"/>
    </div>
</div>
<div class='title'>Test Entry</div>
<p class="p2"><span class="s2">Test body text</span></p>
</body>
</html>`

	entryPath := filepath.Join(entriesDir, "2025-12-15_Test_Entry.html")
	require.NoError(t, os.WriteFile(entryPath, []byte(htmlContent), 0o600))

	resourcePath := filepath.Join(resourcesDir, "TEST-UUID-1234-5678-ABCD.jpg")
	require.NoError(t, os.WriteFile(resourcePath, []byte("fake JPEG image data"), 0o600))

	metaPath := filepath.Join(resourcesDir, "TEST-UUID-1234-5678-ABCD.json")
	metaData := `{"date": 787654321, "placeName": "Sofia, Bulgaria"}`
	require.NoError(t, os.WriteFile(metaPath, []byte(metaData), 0o600))
}

func verifyZipContents(t *testing.T, zipPath string) {
	t.Helper()

	zipReader, err := zip.OpenReader(zipPath)
	require.NoError(t, err)

	defer func() { _ = zipReader.Close() }() //nolint:errcheck // test cleanup

	var (
		hasJSON, hasPhotosDir bool
		jsonFile              *zip.File
	)

	for _, f := range zipReader.File {
		if strings.HasSuffix(f.Name, ".json") {
			hasJSON = true
			jsonFile = f
		}

		if strings.HasPrefix(f.Name, "photos/") {
			hasPhotosDir = true
		}
	}

	require.True(t, hasJSON, "ZIP should contain JSON file")
	require.True(t, hasPhotosDir, "ZIP should contain photos directory")

	verifyJSONContent(t, jsonFile)
}

func verifyJSONContent(t *testing.T, jsonFile *zip.File) {
	t.Helper()

	require.NotNil(t, jsonFile)

	rc, err := jsonFile.Open()
	require.NoError(t, err)

	defer func() { _ = rc.Close() }() //nolint:errcheck // test cleanup

	var export models.DayOneExport

	require.NoError(t, json.NewDecoder(rc).Decode(&export))

	require.Equal(t, "1.0", export.Metadata.Version)
	require.Len(t, export.Entries, 1)
	require.Contains(t, export.Entries[0].Text, "Test Entry")
	require.Len(t, export.Entries[0].Photos, 1)
}

func TestConvertWithVideo(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	outputPath := filepath.Join(tmpDir, "output.zip")

	setupVideoTestData(t, inputDir)

	conv := converter.NewConverter(inputDir, "VideoJournal")
	conv.SetTimeZone("UTC")

	err := conv.Convert(outputPath)

	require.NoError(t, err)
	require.FileExists(t, outputPath)

	verifyVideoZipContents(t, outputPath)
}

func setupVideoTestData(t *testing.T, inputDir string) {
	t.Helper()

	entriesDir := filepath.Join(inputDir, "Entries")
	resourcesDir := filepath.Join(inputDir, "Resources")

	require.NoError(t, os.MkdirAll(entriesDir, 0o750))
	require.NoError(t, os.MkdirAll(resourcesDir, 0o750))

	htmlContent := `<!DOCTYPE html>
<html>
<body>
<div class="pageHeader">Monday, 15 December 2025</div>
<div class="assetGrid">
    <div id="VIDEO-UUID-1234" class="gridItem assetType_video">
        <video src="../Resources/VIDEO-UUID-1234.mov" class="asset_video"></video>
    </div>
</div>
<div class='title'>Video Entry</div>
</body>
</html>`

	entryPath := filepath.Join(entriesDir, "2025-12-15_Video.html")

	require.NoError(t, os.WriteFile(entryPath, []byte(htmlContent), 0o600))

	videoPath := filepath.Join(resourcesDir, "VIDEO-UUID-1234.mov")

	require.NoError(t, os.WriteFile(videoPath, []byte("fake video data"), 0o600))

	metaPath := filepath.Join(resourcesDir, "VIDEO-UUID-1234.json")
	metaData := `{"date": 787654321, "placeName": "Test Location"}`

	require.NoError(t, os.WriteFile(metaPath, []byte(metaData), 0o600))
}

func verifyVideoZipContents(t *testing.T, zipPath string) {
	t.Helper()

	zipReader, err := zip.OpenReader(zipPath)
	require.NoError(t, err)

	defer func() { _ = zipReader.Close() }() //nolint:errcheck // test cleanup

	var hasVideosDir bool

	for _, f := range zipReader.File {
		if strings.HasPrefix(f.Name, "videos/") {
			hasVideosDir = true

			break
		}
	}

	require.True(t, hasVideosDir, "ZIP should contain videos directory")
}

func TestConvertMultipleEntries(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	outputPath := filepath.Join(tmpDir, "output.zip")

	setupMultipleEntriesData(t, inputDir)

	conv := converter.NewConverter(inputDir, "MultiJournal")
	conv.SetTimeZone("America/New_York")

	err := conv.Convert(outputPath)

	require.NoError(t, err)
	require.FileExists(t, outputPath)

	verifyMultipleEntries(t, outputPath)
}

func setupMultipleEntriesData(t *testing.T, inputDir string) {
	t.Helper()

	entriesDir := filepath.Join(inputDir, "Entries")
	resourcesDir := filepath.Join(inputDir, "Resources")

	require.NoError(t, os.MkdirAll(entriesDir, 0o750))
	require.NoError(t, os.MkdirAll(resourcesDir, 0o750))

	for i := 1; i <= 3; i++ {
		htmlContent := `<!DOCTYPE html>
<html>
<body>
<div class="pageHeader">Monday, 15 December 2025</div>
<div class='title'>Entry ` + string(rune('0'+i)) + `</div>
<p class="p2"><span class="s2">Body text ` + string(rune('0'+i)) + `</span></p>
</body>
</html>`

		entryPath := filepath.Join(entriesDir, "2025-12-15_Entry_"+string(rune('0'+i))+".html")

		require.NoError(t, os.WriteFile(entryPath, []byte(htmlContent), 0o600))
	}
}

func verifyMultipleEntries(t *testing.T, zipPath string) {
	t.Helper()

	zipReader, err := zip.OpenReader(zipPath)
	require.NoError(t, err)

	defer func() { _ = zipReader.Close() }() //nolint:errcheck // test cleanup

	var jsonFile *zip.File

	for _, f := range zipReader.File {
		if strings.HasSuffix(f.Name, ".json") {
			jsonFile = f

			break
		}
	}

	require.NotNil(t, jsonFile)

	rc, err := jsonFile.Open()
	require.NoError(t, err)

	defer func() { _ = rc.Close() }() //nolint:errcheck // test cleanup

	var export models.DayOneExport

	require.NoError(t, json.NewDecoder(rc).Decode(&export))
	require.Len(t, export.Entries, 3)
}

func TestConvertWithHEIC(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	outputPath := filepath.Join(tmpDir, "output.zip")

	setupHEICTestData(t, inputDir)

	conv := converter.NewConverter(inputDir, "HEICJournal")

	err := conv.Convert(outputPath)

	require.NoError(t, err)
	require.FileExists(t, outputPath)
}

func setupHEICTestData(t *testing.T, inputDir string) {
	t.Helper()

	entriesDir := filepath.Join(inputDir, "Entries")
	resourcesDir := filepath.Join(inputDir, "Resources")

	require.NoError(t, os.MkdirAll(entriesDir, 0o750))
	require.NoError(t, os.MkdirAll(resourcesDir, 0o750))

	htmlContent := `<!DOCTYPE html>
<html>
<body>
<div class="pageHeader">Monday, 15 December 2025</div>
<div class="assetGrid">
    <div id="HEIC-UUID-1234" class="gridItem assetType_photo">
        <img src="../Resources/HEIC-UUID-1234.heic" class="asset_image"/>
    </div>
</div>
<div class='title'>HEIC Entry</div>
</body>
</html>`

	entryPath := filepath.Join(entriesDir, "2025-12-15_HEIC.html")

	require.NoError(t, os.WriteFile(entryPath, []byte(htmlContent), 0o600))

	heicPath := filepath.Join(resourcesDir, "HEIC-UUID-1234.heic")

	require.NoError(t, os.WriteFile(heicPath, []byte("fake HEIC data"), 0o600))
}

func TestConvertWithMixedMedia(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	outputPath := filepath.Join(tmpDir, "output.zip")

	setupMixedMediaTestData(t, inputDir)

	conv := converter.NewConverter(inputDir, "MixedJournal")

	err := conv.Convert(outputPath)

	require.NoError(t, err)
	require.FileExists(t, outputPath)

	verifyMixedMediaZipContents(t, outputPath)
}

func setupMixedMediaTestData(t *testing.T, inputDir string) {
	t.Helper()

	entriesDir := filepath.Join(inputDir, "Entries")
	resourcesDir := filepath.Join(inputDir, "Resources")

	require.NoError(t, os.MkdirAll(entriesDir, 0o750))
	require.NoError(t, os.MkdirAll(resourcesDir, 0o750))

	htmlContent := `<!DOCTYPE html>
<html>
<body>
<div class="pageHeader">Monday, 15 December 2025</div>
<div class="assetGrid">
    <div id="PHOTO-UUID-1" class="gridItem assetType_photo">
        <img src="../Resources/PHOTO-UUID-1.jpg" class="asset_image"/>
    </div>
    <div id="VIDEO-UUID-1" class="gridItem assetType_video">
        <video src="../Resources/VIDEO-UUID-1.mp4" class="asset_video"></video>
    </div>
    <div id="PHOTO-UUID-2" class="gridItem assetType_photo">
        <img src="../Resources/PHOTO-UUID-2.heic" class="asset_image"/>
    </div>
</div>
<div class='title'>Mixed Media Entry</div>
</body>
</html>`

	entryPath := filepath.Join(entriesDir, "2025-12-15_Mixed.html")

	require.NoError(t, os.WriteFile(entryPath, []byte(htmlContent), 0o600))

	require.NoError(t, os.WriteFile(filepath.Join(resourcesDir, "PHOTO-UUID-1.jpg"), []byte("fake jpg"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(resourcesDir, "VIDEO-UUID-1.mp4"), []byte("fake mp4"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(resourcesDir, "PHOTO-UUID-2.heic"), []byte("fake heic"), 0o600))
}

func verifyMixedMediaZipContents(t *testing.T, zipPath string) {
	t.Helper()

	zipReader, err := zip.OpenReader(zipPath)
	require.NoError(t, err)

	defer func() { _ = zipReader.Close() }() //nolint:errcheck // test cleanup

	var (
		hasPhotos bool
		hasVideos bool
	)

	for _, f := range zipReader.File {
		if strings.HasPrefix(f.Name, "photos/") {
			hasPhotos = true
		}

		if strings.HasPrefix(f.Name, "videos/") {
			hasVideos = true
		}
	}

	require.True(t, hasPhotos, "ZIP should contain photos")
	require.True(t, hasVideos, "ZIP should contain videos")
}

func TestConvertWithEntryWithoutAssets(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	outputPath := filepath.Join(tmpDir, "output.zip")

	setupNoAssetsTestData(t, inputDir)

	conv := converter.NewConverter(inputDir, "NoAssetsJournal")

	err := conv.Convert(outputPath)

	require.NoError(t, err)
	require.FileExists(t, outputPath)
}

func setupNoAssetsTestData(t *testing.T, inputDir string) {
	t.Helper()

	entriesDir := filepath.Join(inputDir, "Entries")
	resourcesDir := filepath.Join(inputDir, "Resources")

	require.NoError(t, os.MkdirAll(entriesDir, 0o750))
	require.NoError(t, os.MkdirAll(resourcesDir, 0o750))

	htmlContent := `<!DOCTYPE html>
<html>
<body>
<div class="pageHeader">Monday, 15 December 2025</div>
<div class='title'>Text Only Entry</div>
<p class="p2"><span class="s2">This entry has no photos or videos.</span></p>
</body>
</html>`

	entryPath := filepath.Join(entriesDir, "2025-12-15_TextOnly.html")

	require.NoError(t, os.WriteFile(entryPath, []byte(htmlContent), 0o600))
}

func TestConvertWithDifferentVideoFormats(t *testing.T) {
	t.Parallel()

	formats := []string{"mov", "mp4", "m4v", "avi"}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			inputDir := filepath.Join(tmpDir, "input")
			outputPath := filepath.Join(tmpDir, "output.zip")

			setupVideoFormatTestData(t, inputDir, format)

			conv := converter.NewConverter(inputDir, "VideoFormatJournal")

			err := conv.Convert(outputPath)

			require.NoError(t, err)
			require.FileExists(t, outputPath)
		})
	}
}

func setupVideoFormatTestData(t *testing.T, inputDir, format string) {
	t.Helper()

	entriesDir := filepath.Join(inputDir, "Entries")
	resourcesDir := filepath.Join(inputDir, "Resources")

	require.NoError(t, os.MkdirAll(entriesDir, 0o750))
	require.NoError(t, os.MkdirAll(resourcesDir, 0o750))

	htmlContent := `<!DOCTYPE html>
<html>
<body>
<div class="pageHeader">Monday, 15 December 2025</div>
<div class="assetGrid">
    <div id="VIDEO-FORMAT-UUID" class="gridItem assetType_video">
        <video src="../Resources/VIDEO-FORMAT-UUID.` + format + `" class="asset_video"></video>
    </div>
</div>
<div class='title'>Video Format Entry</div>
</body>
</html>`

	entryPath := filepath.Join(entriesDir, "2025-12-15_VideoFormat.html")

	require.NoError(t, os.WriteFile(entryPath, []byte(htmlContent), 0o600))

	videoPath := filepath.Join(resourcesDir, "VIDEO-FORMAT-UUID."+format)

	require.NoError(t, os.WriteFile(videoPath, []byte("fake video data"), 0o600))
}

func TestConvertWithMissingResource(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	outputPath := filepath.Join(tmpDir, "output.zip")

	setupMissingResourceTestData(t, inputDir)

	conv := converter.NewConverter(inputDir, "MissingResourceJournal")

	err := conv.Convert(outputPath)

	require.NoError(t, err)
	require.FileExists(t, outputPath)
}

func setupMissingResourceTestData(t *testing.T, inputDir string) {
	t.Helper()

	entriesDir := filepath.Join(inputDir, "Entries")
	resourcesDir := filepath.Join(inputDir, "Resources")

	require.NoError(t, os.MkdirAll(entriesDir, 0o750))
	require.NoError(t, os.MkdirAll(resourcesDir, 0o750))

	htmlContent := `<!DOCTYPE html>
<html>
<body>
<div class="pageHeader">Monday, 15 December 2025</div>
<div class="assetGrid">
    <div id="MISSING-RESOURCE-UUID" class="gridItem assetType_photo">
        <img src="../Resources/MISSING-RESOURCE-UUID.jpg" class="asset_image"/>
    </div>
</div>
<div class='title'>Missing Resource Entry</div>
</body>
</html>`

	entryPath := filepath.Join(entriesDir, "2025-12-15_Missing.html")

	require.NoError(t, os.WriteFile(entryPath, []byte(htmlContent), 0o600))
}

func TestConvertWithResourceMetadata(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	outputPath := filepath.Join(tmpDir, "output.zip")

	setupResourceMetadataTestData(t, inputDir)

	conv := converter.NewConverter(inputDir, "MetadataJournal")
	conv.SetTimeZone("Europe/London")

	err := conv.Convert(outputPath)

	require.NoError(t, err)
	require.FileExists(t, outputPath)

	verifyResourceMetadata(t, outputPath)
}

func setupResourceMetadataTestData(t *testing.T, inputDir string) {
	t.Helper()

	entriesDir := filepath.Join(inputDir, "Entries")
	resourcesDir := filepath.Join(inputDir, "Resources")

	require.NoError(t, os.MkdirAll(entriesDir, 0o750))
	require.NoError(t, os.MkdirAll(resourcesDir, 0o750))

	htmlContent := `<!DOCTYPE html>
<html>
<body>
<div class="pageHeader">Monday, 15 December 2025</div>
<div class="assetGrid">
    <div id="META-UUID-1234" class="gridItem assetType_photo">
        <img src="../Resources/META-UUID-1234.jpg" class="asset_image"/>
    </div>
</div>
<div class='title'>Metadata Entry</div>
</body>
</html>`

	entryPath := filepath.Join(entriesDir, "2025-12-15_Meta.html")

	require.NoError(t, os.WriteFile(entryPath, []byte(htmlContent), 0o600))

	photoPath := filepath.Join(resourcesDir, "META-UUID-1234.jpg")

	require.NoError(t, os.WriteFile(photoPath, []byte("fake photo data"), 0o600))

	metaPath := filepath.Join(resourcesDir, "META-UUID-1234.json")
	metaData := `{"date": 787654321, "placeName": "London, UK"}`

	require.NoError(t, os.WriteFile(metaPath, []byte(metaData), 0o600))
}

func verifyResourceMetadata(t *testing.T, zipPath string) {
	t.Helper()

	zipReader, err := zip.OpenReader(zipPath)
	require.NoError(t, err)

	defer func() { _ = zipReader.Close() }() //nolint:errcheck // test cleanup

	var jsonFile *zip.File

	for _, f := range zipReader.File {
		if strings.HasSuffix(f.Name, ".json") {
			jsonFile = f

			break
		}
	}

	require.NotNil(t, jsonFile)

	rc, err := jsonFile.Open()
	require.NoError(t, err)

	defer func() { _ = rc.Close() }() //nolint:errcheck // test cleanup

	var export models.DayOneExport

	require.NoError(t, json.NewDecoder(rc).Decode(&export))
	require.Len(t, export.Entries, 1)
	require.Len(t, export.Entries[0].Photos, 1)
	require.NotEmpty(t, export.Entries[0].Photos[0].Date)
}

func TestConvertWithJPGExtension(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	outputPath := filepath.Join(tmpDir, "output.zip")

	setupJPGTestData(t, inputDir)

	conv := converter.NewConverter(inputDir, "JPGJournal")

	err := conv.Convert(outputPath)

	require.NoError(t, err)
	require.FileExists(t, outputPath)

	verifyJPGConversion(t, outputPath)
}

func setupJPGTestData(t *testing.T, inputDir string) {
	t.Helper()

	entriesDir := filepath.Join(inputDir, "Entries")
	resourcesDir := filepath.Join(inputDir, "Resources")

	require.NoError(t, os.MkdirAll(entriesDir, 0o750))
	require.NoError(t, os.MkdirAll(resourcesDir, 0o750))

	htmlContent := `<!DOCTYPE html>
<html>
<body>
<div class="pageHeader">Monday, 15 December 2025</div>
<div class="assetGrid">
    <div id="JPG-UUID-1234" class="gridItem assetType_photo">
        <img src="../Resources/JPG-UUID-1234.JPG" class="asset_image"/>
    </div>
</div>
<div class='title'>JPG Entry</div>
</body>
</html>`

	entryPath := filepath.Join(entriesDir, "2025-12-15_JPG.html")

	require.NoError(t, os.WriteFile(entryPath, []byte(htmlContent), 0o600))

	jpgPath := filepath.Join(resourcesDir, "JPG-UUID-1234.JPG")

	require.NoError(t, os.WriteFile(jpgPath, []byte("fake JPG data"), 0o600))
}

func verifyJPGConversion(t *testing.T, zipPath string) {
	t.Helper()

	zipReader, err := zip.OpenReader(zipPath)
	require.NoError(t, err)

	defer func() { _ = zipReader.Close() }() //nolint:errcheck // test cleanup

	var jsonFile *zip.File

	for _, f := range zipReader.File {
		if strings.HasSuffix(f.Name, ".json") {
			jsonFile = f

			break
		}
	}

	require.NotNil(t, jsonFile)

	rc, err := jsonFile.Open()
	require.NoError(t, err)

	defer func() { _ = rc.Close() }() //nolint:errcheck // test cleanup

	var export models.DayOneExport

	require.NoError(t, json.NewDecoder(rc).Decode(&export))
	require.Len(t, export.Entries, 1)
	require.Len(t, export.Entries[0].Photos, 1)
	require.Equal(t, "jpeg", export.Entries[0].Photos[0].Type)
}

func TestConvertWithPNG(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	outputPath := filepath.Join(tmpDir, "output.zip")

	setupPNGTestData(t, inputDir)

	conv := converter.NewConverter(inputDir, "PNGJournal")

	err := conv.Convert(outputPath)

	require.NoError(t, err)
	require.FileExists(t, outputPath)
}

func setupPNGTestData(t *testing.T, inputDir string) {
	t.Helper()

	entriesDir := filepath.Join(inputDir, "Entries")
	resourcesDir := filepath.Join(inputDir, "Resources")

	require.NoError(t, os.MkdirAll(entriesDir, 0o750))
	require.NoError(t, os.MkdirAll(resourcesDir, 0o750))

	htmlContent := `<!DOCTYPE html>
<html>
<body>
<div class="pageHeader">Monday, 15 December 2025</div>
<div class="assetGrid">
    <div id="PNG-UUID-1234" class="gridItem assetType_photo">
        <img src="../Resources/PNG-UUID-1234.png" class="asset_image"/>
    </div>
</div>
<div class='title'>PNG Entry</div>
</body>
</html>`

	entryPath := filepath.Join(entriesDir, "2025-12-15_PNG.html")

	require.NoError(t, os.WriteFile(entryPath, []byte(htmlContent), 0o600))

	pngPath := filepath.Join(resourcesDir, "PNG-UUID-1234.png")

	require.NoError(t, os.WriteFile(pngPath, []byte("fake PNG data"), 0o600))
}

func TestConvertWithMultiplePhotos(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	outputPath := filepath.Join(tmpDir, "output.zip")

	setupMultiplePhotosTestData(t, inputDir)

	conv := converter.NewConverter(inputDir, "MultiPhotoJournal")

	err := conv.Convert(outputPath)

	require.NoError(t, err)
	require.FileExists(t, outputPath)

	verifyMultiplePhotos(t, outputPath)
}

func setupMultiplePhotosTestData(t *testing.T, inputDir string) {
	t.Helper()

	entriesDir := filepath.Join(inputDir, "Entries")
	resourcesDir := filepath.Join(inputDir, "Resources")

	require.NoError(t, os.MkdirAll(entriesDir, 0o750))
	require.NoError(t, os.MkdirAll(resourcesDir, 0o750))

	htmlContent := `<!DOCTYPE html>
<html>
<body>
<div class="pageHeader">Monday, 15 December 2025</div>
<div class="assetGrid">
    <div id="MULTI-PHOTO-1" class="gridItem assetType_photo">
        <img src="../Resources/MULTI-PHOTO-1.jpg" class="asset_image"/>
    </div>
    <div id="MULTI-PHOTO-2" class="gridItem assetType_photo">
        <img src="../Resources/MULTI-PHOTO-2.jpg" class="asset_image"/>
    </div>
    <div id="MULTI-PHOTO-3" class="gridItem assetType_photo">
        <img src="../Resources/MULTI-PHOTO-3.jpg" class="asset_image"/>
    </div>
</div>
<div class='title'>Multiple Photos Entry</div>
</body>
</html>`

	entryPath := filepath.Join(entriesDir, "2025-12-15_MultiPhoto.html")

	require.NoError(t, os.WriteFile(entryPath, []byte(htmlContent), 0o600))

	for i := 1; i <= 3; i++ {
		photoPath := filepath.Join(resourcesDir, "MULTI-PHOTO-"+string(rune('0'+i))+".jpg")

		require.NoError(t, os.WriteFile(photoPath, []byte("fake photo data "+string(rune('0'+i))), 0o600))
	}
}

func verifyMultiplePhotos(t *testing.T, zipPath string) {
	t.Helper()

	zipReader, err := zip.OpenReader(zipPath)
	require.NoError(t, err)

	defer func() { _ = zipReader.Close() }() //nolint:errcheck // test cleanup

	var photoCount int

	for _, f := range zipReader.File {
		if strings.HasPrefix(f.Name, "photos/") {
			photoCount++
		}
	}

	require.Equal(t, 3, photoCount)
}

func TestConvertOutputDirError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	outputPath := filepath.Join(tmpDir, "nested", "deep", "output.zip")

	setupConvertTestData(t, inputDir)

	conv := converter.NewConverter(inputDir, "NestedJournal")

	err := conv.Convert(outputPath)

	require.Error(t, err)
}

func TestConvertWithSkippedAssetType(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	outputPath := filepath.Join(tmpDir, "output.zip")

	setupSkippedAssetTestData(t, inputDir)

	conv := converter.NewConverter(inputDir, "SkippedJournal")

	err := conv.Convert(outputPath)

	require.NoError(t, err)
	require.FileExists(t, outputPath)
}

func setupSkippedAssetTestData(t *testing.T, inputDir string) {
	t.Helper()

	entriesDir := filepath.Join(inputDir, "Entries")
	resourcesDir := filepath.Join(inputDir, "Resources")

	require.NoError(t, os.MkdirAll(entriesDir, 0o750))
	require.NoError(t, os.MkdirAll(resourcesDir, 0o750))

	htmlContent := `<!DOCTYPE html>
<html>
<body>
<div class="pageHeader">Monday, 15 December 2025</div>
<div class="assetGrid">
    <div id="PHOTO-UUID" class="gridItem assetType_photo">
        <img src="../Resources/PHOTO-UUID.jpg" class="asset_image"/>
    </div>
    <div id="AUDIO-UUID" class="gridItem assetType_voice">
    </div>
</div>
<div class='title'>Skipped Asset Entry</div>
</body>
</html>`

	entryPath := filepath.Join(entriesDir, "2025-12-15_Skipped.html")

	require.NoError(t, os.WriteFile(entryPath, []byte(htmlContent), 0o600))

	photoPath := filepath.Join(resourcesDir, "PHOTO-UUID.jpg")

	require.NoError(t, os.WriteFile(photoPath, []byte("fake photo"), 0o600))
}

func TestConvertTitleOnly(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	outputPath := filepath.Join(tmpDir, "output.zip")

	setupTitleOnlyTestData(t, inputDir)

	conv := converter.NewConverter(inputDir, "TitleOnlyJournal")

	err := conv.Convert(outputPath)

	require.NoError(t, err)
	require.FileExists(t, outputPath)
}

func setupTitleOnlyTestData(t *testing.T, inputDir string) {
	t.Helper()

	entriesDir := filepath.Join(inputDir, "Entries")
	resourcesDir := filepath.Join(inputDir, "Resources")

	require.NoError(t, os.MkdirAll(entriesDir, 0o750))
	require.NoError(t, os.MkdirAll(resourcesDir, 0o750))

	htmlContent := `<!DOCTYPE html>
<html>
<body>
<div class="pageHeader">Monday, 15 December 2025</div>
<div class='title'>Title Only No Body</div>
</body>
</html>`

	entryPath := filepath.Join(entriesDir, "2025-12-15_TitleOnly.html")

	require.NoError(t, os.WriteFile(entryPath, []byte(htmlContent), 0o600))
}

func TestConvertBodyOnly(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	outputPath := filepath.Join(tmpDir, "output.zip")

	setupBodyOnlyTestData(t, inputDir)

	conv := converter.NewConverter(inputDir, "BodyOnlyJournal")

	err := conv.Convert(outputPath)

	require.NoError(t, err)
	require.FileExists(t, outputPath)
}

func setupBodyOnlyTestData(t *testing.T, inputDir string) {
	t.Helper()

	entriesDir := filepath.Join(inputDir, "Entries")
	resourcesDir := filepath.Join(inputDir, "Resources")

	require.NoError(t, os.MkdirAll(entriesDir, 0o750))
	require.NoError(t, os.MkdirAll(resourcesDir, 0o750))

	htmlContent := `<!DOCTYPE html>
<html>
<body>
<div class="pageHeader">Monday, 15 December 2025</div>
<p class="p2"><span class="s2">Body text without title.</span></p>
</body>
</html>`

	entryPath := filepath.Join(entriesDir, "2025-12-15_BodyOnly.html")

	require.NoError(t, os.WriteFile(entryPath, []byte(htmlContent), 0o600))
}

func TestConvertEmptyEntry(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	outputPath := filepath.Join(tmpDir, "output.zip")

	setupEmptyEntryTestData(t, inputDir)

	conv := converter.NewConverter(inputDir, "EmptyJournal")

	err := conv.Convert(outputPath)

	require.NoError(t, err)
	require.FileExists(t, outputPath)
}

func setupEmptyEntryTestData(t *testing.T, inputDir string) {
	t.Helper()

	entriesDir := filepath.Join(inputDir, "Entries")
	resourcesDir := filepath.Join(inputDir, "Resources")

	require.NoError(t, os.MkdirAll(entriesDir, 0o750))
	require.NoError(t, os.MkdirAll(resourcesDir, 0o750))

	htmlContent := `<!DOCTYPE html>
<html>
<body>
<div class="pageHeader">Monday, 15 December 2025</div>
</body>
</html>`

	entryPath := filepath.Join(entriesDir, "2025-12-15_Empty.html")

	require.NoError(t, os.WriteFile(entryPath, []byte(htmlContent), 0o600))
}

func TestConvertLargeImage(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	outputPath := filepath.Join(tmpDir, "output.zip")

	setupLargeImageTestData(t, inputDir)

	conv := converter.NewConverter(inputDir, "LargeImageJournal")

	err := conv.Convert(outputPath)

	require.NoError(t, err)
	require.FileExists(t, outputPath)
}

func setupLargeImageTestData(t *testing.T, inputDir string) {
	t.Helper()

	entriesDir := filepath.Join(inputDir, "Entries")
	resourcesDir := filepath.Join(inputDir, "Resources")

	require.NoError(t, os.MkdirAll(entriesDir, 0o750))
	require.NoError(t, os.MkdirAll(resourcesDir, 0o750))

	htmlContent := `<!DOCTYPE html>
<html>
<body>
<div class="pageHeader">Monday, 15 December 2025</div>
<div class="assetGrid">
    <div id="LARGE-UUID" class="gridItem assetType_photo">
        <img src="../Resources/LARGE-UUID.jpg" class="asset_image"/>
    </div>
</div>
<div class='title'>Large Image Entry</div>
</body>
</html>`

	entryPath := filepath.Join(entriesDir, "2025-12-15_Large.html")

	require.NoError(t, os.WriteFile(entryPath, []byte(htmlContent), 0o600))

	largeData := make([]byte, 1024*1024)

	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	imgPath := filepath.Join(resourcesDir, "LARGE-UUID.jpg")

	require.NoError(t, os.WriteFile(imgPath, largeData, 0o600))
}
