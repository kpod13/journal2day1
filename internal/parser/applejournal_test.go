package parser_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kpod13/journal2day1/internal/parser"
)

func TestParseEntry(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupTestStructure(t, tmpDir)

	p := parser.NewAppleJournalParser(tmpDir)
	entryPath := filepath.Join(tmpDir, "Entries", "2025-12-15_Test.html")

	entry, err := p.ParseEntry(entryPath)
	require.NoError(t, err)

	require.Equal(t, "Test Entry Title", entry.Title)
	require.Equal(t, time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC), entry.Date)
	require.Len(t, entry.Assets, 1)
	require.Equal(t, "TEST-UUID-1234", entry.Assets[0].ID)
	require.Equal(t, "photo", entry.Assets[0].Type)
}

func TestParseAll(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	setupMultipleEntries(t, tmpDir)

	p := parser.NewAppleJournalParser(tmpDir)

	entries, err := p.ParseAll()
	require.NoError(t, err)
	require.Len(t, entries, 3)
}

func TestParseEntryWithBody(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createDirs(t, tmpDir)

	content := `<!DOCTYPE html>
<html>
<body>
<div class="pageHeader">Monday, 15 December 2025</div>
<div class='title'>Test Title</div>
<div class='bodyText'></div>
<p class="p2"><span class="s2">Test body text</span></p>
</body>
</html>`

	entryPath := filepath.Join(tmpDir, "Entries", "2025-12-15_Test.html")
	require.NoError(t, os.WriteFile(entryPath, []byte(content), 0o600))

	p := parser.NewAppleJournalParser(tmpDir)

	entry, err := p.ParseEntry(entryPath)
	require.NoError(t, err)
	require.Equal(t, "Test Title", entry.Title)
	require.Equal(t, "Test body text", entry.Body)
}

func TestLoadResourceMeta(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createDirs(t, tmpDir)

	metaContent := `{"date": 784043393, "placeName": "Moscow"}`
	metaPath := filepath.Join(tmpDir, "Resources", "test-uuid.json")
	require.NoError(t, os.WriteFile(metaPath, []byte(metaContent), 0o600))

	p := parser.NewAppleJournalParser(tmpDir)

	meta, err := p.LoadResourceMeta("test-uuid")
	require.NoError(t, err)
	require.Equal(t, float64(784043393), meta.Date)
	require.Equal(t, "Moscow", meta.PlaceName)
}

func setupTestStructure(t *testing.T, tmpDir string) {
	t.Helper()

	createDirs(t, tmpDir)
	writeTestEntry(t, tmpDir)
	writeTestResource(t, tmpDir)
}

func setupMultipleEntries(t *testing.T, tmpDir string) {
	t.Helper()

	createDirs(t, tmpDir)

	for i := 1; i <= 3; i++ {
		writeNumberedEntry(t, tmpDir, i)
	}
}

func createDirs(t *testing.T, tmpDir string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "Entries"), 0o750))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "Resources"), 0o750))
}

func writeTestEntry(t *testing.T, tmpDir string) {
	t.Helper()

	content := `<!DOCTYPE html>
<html>
<body>
<div class="pageHeader">Monday, 15 December 2025</div>
<div class="assetGrid">
    <div id="TEST-UUID-1234" class="gridItem assetType_photo">
        <img src="../Resources/TEST-UUID-1234.jpg" class="asset_image"/>
    </div>
</div>
<div class='title'>Test Entry Title</div>
</body>
</html>`

	entryPath := filepath.Join(tmpDir, "Entries", "2025-12-15_Test.html")
	require.NoError(t, os.WriteFile(entryPath, []byte(content), 0o600))
}

func writeTestResource(t *testing.T, tmpDir string) {
	t.Helper()

	resourcePath := filepath.Join(tmpDir, "Resources", "TEST-UUID-1234.jpg")
	require.NoError(t, os.WriteFile(resourcePath, []byte("fake image"), 0o600))
}

func writeNumberedEntry(t *testing.T, tmpDir string, num int) {
	t.Helper()

	content := `<!DOCTYPE html>
<html>
<body>
<div class="pageHeader">Monday, 15 December 2025</div>
<div class='title'>Entry</div>
</body>
</html>`

	filename := filepath.Join(tmpDir, "Entries", "2025-12-15_Entry_"+string(rune('0'+num))+".html")
	require.NoError(t, os.WriteFile(filename, []byte(content), 0o600))
}

func TestGetResourceFilePath(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createDirs(t, tmpDir)

	resourcePath := filepath.Join(tmpDir, "Resources", "TEST-UUID-5678.jpg")
	require.NoError(t, os.WriteFile(resourcePath, []byte("fake image"), 0o600))

	p := parser.NewAppleJournalParser(tmpDir)

	path := p.GetResourceFilePath("TEST-UUID-5678")

	require.NotEmpty(t, path)
	require.Contains(t, path, "TEST-UUID-5678.jpg")
}

func TestGetResourceFilePathNotFound(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createDirs(t, tmpDir)

	p := parser.NewAppleJournalParser(tmpDir)

	path := p.GetResourceFilePath("NONEXISTENT-UUID")

	require.Empty(t, path)
}

func TestParseEntryWithVideo(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createDirs(t, tmpDir)

	content := `<!DOCTYPE html>
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

	entryPath := filepath.Join(tmpDir, "Entries", "2025-12-15_Video.html")
	require.NoError(t, os.WriteFile(entryPath, []byte(content), 0o600))

	videoPath := filepath.Join(tmpDir, "Resources", "VIDEO-UUID-1234.mov")
	require.NoError(t, os.WriteFile(videoPath, []byte("fake video"), 0o600))

	p := parser.NewAppleJournalParser(tmpDir)

	entry, err := p.ParseEntry(entryPath)

	require.NoError(t, err)
	require.Equal(t, "Video Entry", entry.Title)
	require.Len(t, entry.Assets, 1)
	require.Equal(t, "VIDEO-UUID-1234", entry.Assets[0].ID)
	require.Equal(t, "video", entry.Assets[0].Type)
}

func TestParseEntryWithVideoSource(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createDirs(t, tmpDir)

	content := `<!DOCTYPE html>
<html>
<body>
<div class="pageHeader">Monday, 15 December 2025</div>
<div class="assetGrid">
    <div id="VIDEO-SOURCE-UUID" class="gridItem assetType_video">
        <video class="asset_video">
            <source src="../Resources/VIDEO-SOURCE-UUID.mp4" type="video/mp4"/>
        </video>
    </div>
</div>
<div class='title'>Video Source Entry</div>
</body>
</html>`

	entryPath := filepath.Join(tmpDir, "Entries", "2025-12-15_VideoSource.html")
	require.NoError(t, os.WriteFile(entryPath, []byte(content), 0o600))

	videoPath := filepath.Join(tmpDir, "Resources", "VIDEO-SOURCE-UUID.mp4")
	require.NoError(t, os.WriteFile(videoPath, []byte("fake video"), 0o600))

	p := parser.NewAppleJournalParser(tmpDir)

	entry, err := p.ParseEntry(entryPath)

	require.NoError(t, err)
	require.Equal(t, "Video Source Entry", entry.Title)
	require.Len(t, entry.Assets, 1)
	require.Equal(t, "VIDEO-SOURCE-UUID", entry.Assets[0].ID)
}

func TestLoadResourceMetaNotFound(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createDirs(t, tmpDir)

	p := parser.NewAppleJournalParser(tmpDir)

	_, err := p.LoadResourceMeta("nonexistent-uuid")

	require.Error(t, err)
}

func TestLoadResourceMetaInvalidJSON(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createDirs(t, tmpDir)

	metaPath := filepath.Join(tmpDir, "Resources", "invalid-uuid.json")
	require.NoError(t, os.WriteFile(metaPath, []byte("invalid json"), 0o600))

	p := parser.NewAppleJournalParser(tmpDir)

	_, err := p.LoadResourceMeta("invalid-uuid")

	require.Error(t, err)
}

func TestParseEntryWithMultipleBodyParagraphs(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createDirs(t, tmpDir)

	content := `<!DOCTYPE html>
<html>
<body>
<div class="pageHeader">Monday, 15 December 2025</div>
<div class='title'>Multi Body</div>
<p class="p2"><span class="s2">First paragraph</span></p>
<p class="p2"><span class="s2">Second paragraph</span></p>
<p class="p2"><span class="s2">Third paragraph</span></p>
</body>
</html>`

	entryPath := filepath.Join(tmpDir, "Entries", "2025-12-15_Multi.html")
	require.NoError(t, os.WriteFile(entryPath, []byte(content), 0o600))

	p := parser.NewAppleJournalParser(tmpDir)

	entry, err := p.ParseEntry(entryPath)

	require.NoError(t, err)
	require.Contains(t, entry.Body, "First paragraph")
	require.Contains(t, entry.Body, "Second paragraph")
	require.Contains(t, entry.Body, "Third paragraph")
}

func TestParseEntryWithBodyTextDiv(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createDirs(t, tmpDir)

	content := `<!DOCTYPE html>
<html>
<body>
<div class="pageHeader">Monday, 15 December 2025</div>
<div class='title'>Body Text Div</div>
<div class='bodyText'>
    <span>This is body text from div</span>
</div>
</body>
</html>`

	entryPath := filepath.Join(tmpDir, "Entries", "2025-12-15_BodyDiv.html")
	require.NoError(t, os.WriteFile(entryPath, []byte(content), 0o600))

	p := parser.NewAppleJournalParser(tmpDir)

	entry, err := p.ParseEntry(entryPath)

	require.NoError(t, err)
	require.Contains(t, entry.Body, "This is body text from div")
}

func TestParseEntryWithDifferentDateFormats(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		header   string
		expected time.Time
	}{
		{
			name:     "full format with day name",
			header:   "Monday, 15 December 2025",
			expected: time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "format without day name",
			header:   "15 December 2025",
			expected: time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "single digit day",
			header:   "Monday, 5 January 2025",
			expected: time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			createDirs(t, tmpDir)

			content := `<!DOCTYPE html>
<html>
<body>
<div class="pageHeader">` + tc.header + `</div>
<div class='title'>Date Test</div>
</body>
</html>`

			entryPath := filepath.Join(tmpDir, "Entries", "test.html")
			require.NoError(t, os.WriteFile(entryPath, []byte(content), 0o600))

			p := parser.NewAppleJournalParser(tmpDir)

			entry, err := p.ParseEntry(entryPath)

			require.NoError(t, err)
			require.Equal(t, tc.expected, entry.Date)
		})
	}
}

func TestParseEntryWithNoAssetFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createDirs(t, tmpDir)

	content := `<!DOCTYPE html>
<html>
<body>
<div class="pageHeader">Monday, 15 December 2025</div>
<div class="assetGrid">
    <div id="MISSING-UUID" class="gridItem assetType_photo">
    </div>
</div>
<div class='title'>Missing Asset</div>
</body>
</html>`

	entryPath := filepath.Join(tmpDir, "Entries", "2025-12-15_Missing.html")
	require.NoError(t, os.WriteFile(entryPath, []byte(content), 0o600))

	p := parser.NewAppleJournalParser(tmpDir)

	entry, err := p.ParseEntry(entryPath)

	require.NoError(t, err)
	require.Equal(t, "Missing Asset", entry.Title)
}

func TestParseAllWithEmptyDirectory(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createDirs(t, tmpDir)

	p := parser.NewAppleJournalParser(tmpDir)

	entries, err := p.ParseAll()

	require.NoError(t, err)
	require.Empty(t, entries)
}

func TestParseEntryFileNotFound(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createDirs(t, tmpDir)

	p := parser.NewAppleJournalParser(tmpDir)

	_, err := p.ParseEntry(filepath.Join(tmpDir, "Entries", "nonexistent.html"))

	require.Error(t, err)
}

func TestParseEntryWithDateFromAssets(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createDirs(t, tmpDir)

	content := `<!DOCTYPE html>
<html>
<body>
<div class="assetGrid">
    <div id="DATE-FROM-ASSET-UUID" class="gridItem assetType_photo">
        <img src="../Resources/DATE-FROM-ASSET-UUID.jpg" class="asset_image"/>
    </div>
</div>
<div class='title'>Date From Asset</div>
</body>
</html>`

	entryPath := filepath.Join(tmpDir, "Entries", "2025-12-15_DateAsset.html")
	require.NoError(t, os.WriteFile(entryPath, []byte(content), 0o600))

	resourcePath := filepath.Join(tmpDir, "Resources", "DATE-FROM-ASSET-UUID.jpg")
	require.NoError(t, os.WriteFile(resourcePath, []byte("fake image"), 0o600))

	metaPath := filepath.Join(tmpDir, "Resources", "DATE-FROM-ASSET-UUID.json")
	metaData := `{"date": 787654321, "placeName": "Test"}`
	require.NoError(t, os.WriteFile(metaPath, []byte(metaData), 0o600))

	p := parser.NewAppleJournalParser(tmpDir)

	entry, err := p.ParseEntry(entryPath)

	require.NoError(t, err)
	require.Equal(t, "Date From Asset", entry.Title)
	require.False(t, entry.Date.IsZero())
}

func TestParseEntryWithDataImageSrc(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createDirs(t, tmpDir)

	content := `<!DOCTYPE html>
<html>
<body>
<div class="pageHeader">Monday, 15 December 2025</div>
<div class="assetGrid">
    <div id="DATA-IMG-UUID" class="gridItem assetType_photo">
        <img src="data:image/png;base64,iVBORw0KGgo=" class="asset_image"/>
    </div>
</div>
<div class='title'>Data Image</div>
</body>
</html>`

	entryPath := filepath.Join(tmpDir, "Entries", "2025-12-15_DataImg.html")
	require.NoError(t, os.WriteFile(entryPath, []byte(content), 0o600))

	p := parser.NewAppleJournalParser(tmpDir)

	entry, err := p.ParseEntry(entryPath)

	require.NoError(t, err)
	require.Equal(t, "Data Image", entry.Title)
}

func TestParseEntryFindResourceFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createDirs(t, tmpDir)

	content := `<!DOCTYPE html>
<html>
<body>
<div class="pageHeader">Monday, 15 December 2025</div>
<div class="assetGrid">
    <div id="FIND-RESOURCE-UUID" class="gridItem assetType_photo">
    </div>
</div>
<div class='title'>Find Resource</div>
</body>
</html>`

	entryPath := filepath.Join(tmpDir, "Entries", "2025-12-15_FindRes.html")
	require.NoError(t, os.WriteFile(entryPath, []byte(content), 0o600))

	resourcePath := filepath.Join(tmpDir, "Resources", "FIND-RESOURCE-UUID.heic")
	require.NoError(t, os.WriteFile(resourcePath, []byte("fake heic"), 0o600))

	p := parser.NewAppleJournalParser(tmpDir)

	entry, err := p.ParseEntry(entryPath)

	require.NoError(t, err)
	require.Equal(t, "Find Resource", entry.Title)
	require.Len(t, entry.Assets, 1)
	require.Equal(t, "FIND-RESOURCE-UUID", entry.Assets[0].ID)
	require.Equal(t, "heic", entry.Assets[0].Extension)
}

func TestParseEntryWithOtherParagraphClass(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createDirs(t, tmpDir)

	content := `<!DOCTYPE html>
<html>
<body>
<div class="pageHeader">Monday, 15 December 2025</div>
<div class='title'>Other Class</div>
<p class="p1"><span class="s1">Ignored text</span></p>
<p class="p2"><span class="s2">Included text</span></p>
</body>
</html>`

	entryPath := filepath.Join(tmpDir, "Entries", "2025-12-15_OtherClass.html")
	require.NoError(t, os.WriteFile(entryPath, []byte(content), 0o600))

	p := parser.NewAppleJournalParser(tmpDir)

	entry, err := p.ParseEntry(entryPath)

	require.NoError(t, err)
	require.Equal(t, "Included text", entry.Body)
	require.NotContains(t, entry.Body, "Ignored text")
}

func TestParseAllSkipsNonHTML(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createDirs(t, tmpDir)

	htmlContent := `<!DOCTYPE html>
<html>
<body>
<div class="pageHeader">Monday, 15 December 2025</div>
<div class='title'>Real Entry</div>
</body>
</html>`

	entryPath := filepath.Join(tmpDir, "Entries", "2025-12-15_Real.html")
	require.NoError(t, os.WriteFile(entryPath, []byte(htmlContent), 0o600))

	txtPath := filepath.Join(tmpDir, "Entries", "not_html.txt")
	require.NoError(t, os.WriteFile(txtPath, []byte("text file"), 0o600))

	p := parser.NewAppleJournalParser(tmpDir)

	entries, err := p.ParseAll()

	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, "Real Entry", entries[0].Title)
}

func TestParseEntryWithEmptyParagraph(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createDirs(t, tmpDir)

	content := `<!DOCTYPE html>
<html>
<body>
<div class="pageHeader">Monday, 15 December 2025</div>
<div class='title'>Empty Para</div>
<p class="p2"><span class="s2"></span></p>
<p class="p2"><span class="s2">Non-empty</span></p>
</body>
</html>`

	entryPath := filepath.Join(tmpDir, "Entries", "2025-12-15_EmptyPara.html")
	require.NoError(t, os.WriteFile(entryPath, []byte(content), 0o600))

	p := parser.NewAppleJournalParser(tmpDir)

	entry, err := p.ParseEntry(entryPath)

	require.NoError(t, err)
	require.Equal(t, "Non-empty", entry.Body)
}
