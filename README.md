# journal2day1

Convert Apple Journal HTML exports to DayOne JSON ZIP format.

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

## DayOne Import

After conversion, import the ZIP file into DayOne:

1. Open DayOne
1. Go to File > Import > DayOne JSON (.zip)
1. Select the generated ZIP file
1. Choose the target journal

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
