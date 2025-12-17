package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRootCmd(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	cmd := newRootCmd(&buf)

	require.Equal(t, "journal2day1", cmd.Use)
	require.NotNil(t, cmd.Commands())
	require.GreaterOrEqual(t, len(cmd.Commands()), 2)
}

func TestNewVersionCmd(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	cmd := newVersionCmd(&buf)
	cmd.Run(cmd, nil)

	output := buf.String()

	require.Contains(t, output, "journal2day1")
	require.Contains(t, output, "commit:")
	require.Contains(t, output, "built:")
}

func TestPrintVersion(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	printVersion(&buf)

	output := buf.String()

	require.Contains(t, output, version)
	require.Contains(t, output, commit)
	require.Contains(t, output, date)
}

func TestValidateInputDir(t *testing.T) {
	t.Parallel()

	t.Run("valid directory", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "Entries"), 0o750))
		require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "Resources"), 0o750))

		err := validateInputDir(tmpDir)

		require.NoError(t, err)
	})

	t.Run("missing Entries", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "Resources"), 0o750))

		err := validateInputDir(tmpDir)

		require.Error(t, err)
		require.ErrorIs(t, err, errMissingEntries)
	})

	t.Run("missing Resources", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "Entries"), 0o750))

		err := validateInputDir(tmpDir)

		require.Error(t, err)
		require.ErrorIs(t, err, errMissingResources)
	})
}

func TestPrintConvertInfo(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	printConvertInfo(&buf, "/input/path", "/output/path", "MyJournal", "Europe/London")

	output := buf.String()

	require.Contains(t, output, "/input/path")
	require.Contains(t, output, "/output/path")
	require.Contains(t, output, "MyJournal")
	require.Contains(t, output, "Europe/London")
}

func TestNewConvertCmd(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	cfg := &appConfig{output: &buf}

	cmd := newConvertCmd(cfg)

	require.Equal(t, "convert", cmd.Use)

	inputFlag := cmd.Flags().Lookup("input")

	require.NotNil(t, inputFlag)
	require.Equal(t, "i", inputFlag.Shorthand)

	outputFlag := cmd.Flags().Lookup("output")

	require.NotNil(t, outputFlag)
	require.Equal(t, "o", outputFlag.Shorthand)

	nameFlag := cmd.Flags().Lookup("name")

	require.NotNil(t, nameFlag)
	require.Equal(t, "Journal", nameFlag.DefValue)

	tzFlag := cmd.Flags().Lookup("timezone")

	require.NotNil(t, tzFlag)
	require.Equal(t, "Europe/Sofia", tzFlag.DefValue)
}

func TestRunConvert(t *testing.T) {
	t.Parallel()

	t.Run("successful conversion", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		inputDir := filepath.Join(tmpDir, "input")
		outputPath := filepath.Join(tmpDir, "output.zip")

		setupTestData(t, inputDir)

		var buf bytes.Buffer
		cfg := &appConfig{
			inputPath:   inputDir,
			outputPath:  outputPath,
			journalName: "Test",
			timeZone:    "UTC",
			output:      &buf,
		}

		err := runConvert(cfg)

		require.NoError(t, err)
		require.FileExists(t, outputPath)
		require.Contains(t, buf.String(), "Conversion completed successfully!")
	})

	t.Run("invalid input directory", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		cfg := &appConfig{
			inputPath:   "/nonexistent/path",
			outputPath:  "/tmp/output.zip",
			journalName: "Test",
			timeZone:    "UTC",
			output:      &buf,
		}

		err := runConvert(cfg)

		require.Error(t, err)
	})
}

func TestRootCmdExecute(t *testing.T) {
	t.Parallel()

	t.Run("help command", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		cmd := newRootCmd(&buf)
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--help"})

		err := cmd.Execute()

		require.NoError(t, err)
		require.Contains(t, buf.String(), "journal2day1")
	})

	t.Run("version command", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer

		cmd := newRootCmd(&buf)
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"version"})

		err := cmd.Execute()

		require.NoError(t, err)

		output := buf.String()

		require.True(t, strings.Contains(output, "journal2day1") || strings.Contains(output, version))
	})
}

func setupTestData(t *testing.T, inputDir string) {
	t.Helper()

	entriesDir := filepath.Join(inputDir, "Entries")
	resourcesDir := filepath.Join(inputDir, "Resources")

	require.NoError(t, os.MkdirAll(entriesDir, 0o750))
	require.NoError(t, os.MkdirAll(resourcesDir, 0o750))

	htmlContent := `<!DOCTYPE html>
<html>
<body>
<div class="pageHeader">Monday, 15 December 2025</div>
<div class='title'>Test Entry</div>
<p class="p2"><span class="s2">Test body</span></p>
</body>
</html>`

	entryPath := filepath.Join(entriesDir, "2025-12-15_Test.html")

	require.NoError(t, os.WriteFile(entryPath, []byte(htmlContent), 0o600))
}

func TestConvertCommandMissingInput(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	cmd := newRootCmd(&buf)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"convert", "-o", "/tmp/out.zip"})

	err := cmd.Execute()

	require.Error(t, err)
}

func TestConvertCommandMissingOutput(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	cmd := newRootCmd(&buf)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"convert", "-i", "/tmp/input"})

	err := cmd.Execute()

	require.Error(t, err)
}

func TestConvertCommandWithCustomName(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "input")
	outputPath := filepath.Join(tmpDir, "output.zip")

	setupTestData(t, inputDir)

	var buf bytes.Buffer

	cmd := newRootCmd(&buf)
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"convert", "-i", inputDir, "-o", outputPath, "-n", "CustomJournal", "-t", "America/New_York"})

	err := cmd.Execute()

	require.NoError(t, err)
	require.FileExists(t, outputPath)

	output := buf.String()

	require.Contains(t, output, "CustomJournal")
	require.Contains(t, output, "America/New_York")
}

func TestValidateInputDirBothMissing(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	err := validateInputDir(tmpDir)

	require.Error(t, err)
	require.ErrorIs(t, err, errMissingEntries)
}

func TestRunConvertValidationError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	var buf bytes.Buffer
	cfg := &appConfig{
		inputPath:   tmpDir,
		outputPath:  filepath.Join(tmpDir, "output.zip"),
		journalName: "Test",
		timeZone:    "UTC",
		output:      &buf,
	}

	err := runConvert(cfg)

	require.Error(t, err)
	require.ErrorIs(t, err, errMissingEntries)
}

func TestConvertCommandInvalidInputDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.zip")

	var buf bytes.Buffer

	cmd := newRootCmd(&buf)
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"convert", "-i", tmpDir, "-o", outputPath})

	err := cmd.Execute()

	require.Error(t, err)
}
