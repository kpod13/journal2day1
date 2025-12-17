// Package main provides the CLI entry point for journal2day1.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"github.com/kpod13/journal2day1/internal/converter"
	"github.com/kpod13/journal2day1/internal/logger"
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
	log         *logger.Logger
}

func newRootCmd(output io.Writer) *cobra.Command {
	cfg := &appConfig{
		output: output,
		log:    logger.New(output),
	}

	rootCmd := &cobra.Command{
		Use:   "journal2day1",
		Short: "Convert Apple Journal exports to DayOne format",
		Long: logger.Bold("journal2day1") + " converts Apple Journal HTML exports to DayOne JSON ZIP format.\n\n" +
			"The tool reads the Apple Journal export directory (containing " +
			logger.Cyan("Entries/") + " and " + logger.Cyan("Resources/") +
			"\nsubdirectories) and creates a ZIP archive that can be imported into DayOne.\n\n" +
			logger.Dim("Example:") + "\n" +
			"  " + logger.Green("journal2day1 convert -i ~/AppleJournalEntries -o ~/dayone-import.zip"),
	}

	rootCmd.SetUsageTemplate(coloredUsageTemplate())
	rootCmd.AddCommand(newConvertCmd(cfg))
	rootCmd.AddCommand(newVersionCmd(cfg.log))

	return rootCmd
}

func coloredUsageTemplate() string {
	return logger.Bold("Usage:") + `{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

` + logger.Bold("Aliases:") + `
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

` + logger.Bold("Examples:") + `
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

` + logger.Bold("Available Commands:") + `{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  ` + logger.Cyan("{{rpad .Name .NamePadding }}") + ` {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  ` + logger.Cyan("{{rpad .Name .NamePadding }}") + ` {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

` + logger.Bold("Additional Commands:") + `{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  ` + logger.Cyan("{{rpad .Name .NamePadding }}") + ` {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

` + logger.Bold("Flags:") + `
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

` + logger.Bold("Global Flags:") + `
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

` + logger.Bold("Additional help topics:") + `{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
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

func newVersionCmd(log *logger.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(_ *cobra.Command, _ []string) {
			printVersion(log)
		},
	}
}

func printVersion(log *logger.Logger) {
	log.Bold("journal2day1 ")
	log.Println("%s", version)
	log.KeyValue("commit", commit)
	log.KeyValue("built", date)
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

	printConvertInfo(cfg.log, absInput, absOutput, cfg.journalName, cfg.timeZone)

	conv := converter.NewConverter(absInput, cfg.journalName)
	conv.SetTimeZone(cfg.timeZone)

	var bar *progressbar.ProgressBar

	conv.SetProgressFunc(func(current, total int) {
		if bar == nil {
			bar = progressbar.NewOptions(total,
				progressbar.OptionSetWriter(cfg.output),
				progressbar.OptionEnableColorCodes(true),
				progressbar.OptionShowCount(),
				progressbar.OptionSetWidth(getProgressBarWidth()),
				progressbar.OptionSetTheme(progressbar.Theme{
					Saucer:        "[green]█[reset]",
					SaucerHead:    "[green]█[reset]",
					SaucerPadding: "░",
					BarStart:      "[",
					BarEnd:        "]",
				}),
				progressbar.OptionOnCompletion(func() {
					fmt.Fprintln(cfg.output)
				}),
			)
		}

		_ = bar.Set(current) //nolint:errcheck // progress bar errors are not critical
	})

	if err := conv.Convert(absOutput); err != nil {
		return errors.Wrap(err, "failed to convert")
	}

	cfg.log.Success("Conversion completed successfully!")

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

func printConvertInfo(log *logger.Logger, input, output, journalName, timeZone string) {
	log.Header("Journal Conversion")
	log.KeyValue("Input", input)
	log.KeyValue("Output", output)
	log.KeyValue("Journal", journalName)
	log.KeyValue("Timezone", timeZone)
	log.Println("")
}

const (
	maxProgressLineWidth = 120
	minProgressWidth     = 20
	progressBarOffset    = 35 // space for percentage, count, time, etc.
)

func getProgressBarWidth() int {
	barWidth := maxProgressLineWidth - progressBarOffset
	if barWidth < minProgressWidth {
		barWidth = minProgressWidth
	}

	return barWidth
}
