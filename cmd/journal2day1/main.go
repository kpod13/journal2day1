// Package main provides the CLI entry point for journal2day1.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kpod13/journal2day1/internal/converter"
)

// Build-time variables.
var (
	version = "dev"     //nolint:gochecknoglobals // set by ldflags
	commit  = "none"    //nolint:gochecknoglobals // set by ldflags
	date    = "unknown" //nolint:gochecknoglobals // set by ldflags
)

// Sentinel errors for validation.
var (
	errMissingEntries   = errors.New("input directory does not contain Entries subdirectory")
	errMissingResources = errors.New("input directory does not contain Resources subdirectory")
)

func main() {
	if err := newRootCmd(os.Stdout).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type appConfig struct {
	inputPath   string
	outputPath  string
	journalName string
	timeZone    string
	output      io.Writer
}

func newRootCmd(output io.Writer) *cobra.Command {
	cfg := &appConfig{output: output}

	rootCmd := &cobra.Command{
		Use:   "journal2day1",
		Short: "Convert Apple Journal exports to DayOne format",
		Long: `journal2day1 converts Apple Journal HTML exports to DayOne JSON ZIP format.

The tool reads the Apple Journal export directory (containing Entries/ and Resources/
subdirectories) and creates a ZIP archive that can be imported into DayOne.

Example:
  journal2day1 convert -i ~/AppleJournalEntries -o ~/dayone-import.zip`,
	}

	rootCmd.AddCommand(newConvertCmd(cfg))
	rootCmd.AddCommand(newVersionCmd(output))

	return rootCmd
}

func newConvertCmd(cfg *appConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "convert",
		Short: "Convert Apple Journal export to DayOne format",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runConvert(cfg)
		},
	}

	cmd.Flags().StringVarP(&cfg.inputPath, "input", "i", "", "Path to Apple Journal export directory (required)")
	cmd.Flags().StringVarP(&cfg.outputPath, "output", "o", "", "Path to output ZIP file (required)")
	cmd.Flags().StringVarP(&cfg.journalName, "name", "n", "Journal", "Name of the journal in DayOne")
	cmd.Flags().StringVarP(&cfg.timeZone, "timezone", "t", "Europe/Sofia", "Timezone for entries")

	if err := cmd.MarkFlagRequired("input"); err != nil {
		panic(fmt.Sprintf("failed to mark input flag required: %v", err))
	}

	if err := cmd.MarkFlagRequired("output"); err != nil {
		panic(fmt.Sprintf("failed to mark output flag required: %v", err))
	}

	return cmd
}

func newVersionCmd(output io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(_ *cobra.Command, _ []string) {
			printVersion(output)
		},
	}
}

func printVersion(w io.Writer) {
	fmt.Fprintf(w, "journal2day1 %s\n", version)
	fmt.Fprintf(w, "  commit: %s\n", commit)
	fmt.Fprintf(w, "  built:  %s\n", date)
}

func runConvert(cfg *appConfig) error {
	absInput, err := filepath.Abs(cfg.inputPath)
	if err != nil {
		return errors.Wrap(err, "failed to resolve input path")
	}

	if err := validateInputDir(absInput); err != nil {
		return err
	}

	absOutput, err := filepath.Abs(cfg.outputPath)
	if err != nil {
		return errors.Wrap(err, "failed to resolve output path")
	}

	printConvertInfo(cfg.output, absInput, absOutput, cfg.journalName, cfg.timeZone)

	conv := converter.NewConverter(absInput, cfg.journalName)
	conv.SetTimeZone(cfg.timeZone)

	if err := conv.Convert(absOutput); err != nil {
		return errors.Wrap(err, "failed to convert")
	}

	fmt.Fprintln(cfg.output, "Conversion completed successfully!")

	return nil
}

func validateInputDir(absInput string) error {
	entriesDir := filepath.Join(absInput, "Entries")
	if _, err := os.Stat(entriesDir); os.IsNotExist(err) {
		return errors.Wrapf(errMissingEntries, "%s", absInput)
	}

	resourcesDir := filepath.Join(absInput, "Resources")
	if _, err := os.Stat(resourcesDir); os.IsNotExist(err) {
		return errors.Wrapf(errMissingResources, "%s", absInput)
	}

	return nil
}

func printConvertInfo(w io.Writer, input, output, journalName, timeZone string) {
	fmt.Fprintf(w, "Converting Apple Journal export from: %s\n", input)
	fmt.Fprintf(w, "Output will be written to: %s\n", output)
	fmt.Fprintf(w, "Journal name: %s\n", journalName)
	fmt.Fprintf(w, "Timezone: %s\n", timeZone)
}
