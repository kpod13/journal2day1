# journal2day1

A cross-platform CLI tool that converts Apple Journal HTML exports into a Day One-compatible format ready for import.

## Why?

Apple Journal is a great app for quick daily journaling, but it lacks advanced
features like search, tags, multiple journals, and cross-platform sync. DayOne
offers all of this and more.

Unfortunately, there's no direct migration path between these apps. Apple
Journal only exports to HTML, while DayOne imports its own JSON format.

**journal2day1** bridges this gap — it converts your Apple Journal export into
a DayOne-compatible ZIP archive, preserving:

- Entry text and titles
- Photos and videos
- Original creation dates
- Media metadata

## How It Works

```text
┌─────────────────┐      ┌─────────────────┐      ┌─────────────────┐
│  Apple Journal  │      │  journal2day1   │      │     DayOne      │
│     (macOS)     │      │      (CLI)      │      │  (macOS/iOS)    │
└────────┬────────┘      └────────┬────────┘      └────────┬────────┘
         │                        │                        │
         │  File > Export...      │                        │
         │                        │                        │
         ▼                        │                        │
┌─────────────────┐               │                        │
│   Export Dir    │               │                        │
│ ┌─────────────┐ │    convert    │                        │
│ │  Entries/   │ │───────────────▶                        │
│ │   *.html    │ │               │                        │
│ ├─────────────┤ │               ▼                        │
│ │ Resources/  │ │      ┌─────────────────┐               │
│ │ *.jpg *.mov │ │      │   output.zip    │  File > Import│
│ └─────────────┘ │      │ ┌─────────────┐ │───────────────▶
└─────────────────┘      │ │ Journal.json│ │               │
                         │ ├─────────────┤ │               │
                         │ │  photos/    │ │               │
                         │ │  videos/    │ │               │
                         │ └─────────────┘ │               │
                         └─────────────────┘               │
                                                           ▼
                                                  ┌─────────────────┐
                                                  │  Your entries   │
                                                  │  in DayOne      │
                                                  └─────────────────┘
```

## Installation

```bash
go install github.com/kpod13/journal2day1/cmd/journal2day1@latest
```

Or build from source:

```bash
git clone https://github.com/kpod13/journal2day1.git
cd journal2day1
make build
```

## Usage

```bash
journal2day1 convert -i ~/AppleJournalEntries -o ~/dayone-import.zip
```

### Options

| Flag         | Short | Description                            | Default        |
| ------------ | ----- | -------------------------------------- | -------------- |
| `--input`    | `-i`  | Path to Apple Journal export directory | (required)     |
| `--output`   | `-o`  | Path to output ZIP file                | (required)     |
| `--name`     | `-n`  | Name of the journal in DayOne          | `Journal`      |
| `--timezone` | `-t`  | Timezone for entries                   | `Europe/Sofia` |

### Example

```bash
journal2day1 convert \
  -i ~/AppleJournalEntries \
  -o ~/Desktop/dayone-import.zip \
  -n "My Journal" \
  -t "America/New_York"
```

## Exporting from Apple Journal (macOS)

To export your entries from Apple Journal:

1. Open the **Journal** app on your Mac (requires macOS Sonoma 14.2 or later)
2. Click **File** in the menu bar
3. Select **Export Journal...**
4. Choose a destination folder (e.g., `~/AppleJournalEntries`)
5. Click **Export**

The export will create a folder containing:

- `Entries/` — HTML files with your journal entries
- `Resources/` — media files (photos, videos) and their metadata

> **Note:** The export includes all entries from your Journal. There is no option
> to export a specific date range.

## Apple Journal Export Structure

The tool expects an Apple Journal export directory with the following structure:

```text
AppleJournalEntries/
├── Entries/
│   ├── 2024-01-15_Entry1.html
│   ├── 2024-01-16_Entry2.html
│   └── ...
└── Resources/
    ├── uuid1.json
    ├── uuid1.jpg
    ├── uuid2.json
    ├── uuid2.mov
    └── ...
```

## Importing into DayOne

After running the conversion, import the generated ZIP file into DayOne:

### On macOS

1. Open **DayOne** on your Mac
2. Click **File** in the menu bar
3. Select **Import** → **DayOne JSON (.zip)**
4. Navigate to your generated ZIP file and select it
5. Choose the target journal or create a new one
6. Click **Import**

### On iOS/iPadOS

1. Transfer the ZIP file to your device (via AirDrop, iCloud Drive, or Files app)
2. Open **DayOne** on your device
3. Tap the **gear icon** (Settings) in the bottom navigation
4. Scroll down and tap **Import**
5. Select **DayOne JSON (.zip)**
6. Navigate to your ZIP file and select it
7. Choose the target journal

> **Tip:** For large imports with many photos/videos, importing on macOS is
> recommended as it handles large files more reliably.

## Development

```bash
# Run all checks
make all

# Run tests
make test

# Run linter
make lint

# Build binary
make build
```

## License

MIT
