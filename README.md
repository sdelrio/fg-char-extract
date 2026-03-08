# Fantasy Grounds Character Extractor

A Go tool to extract character sheets from a Fantasy Grounds `db.xml` file into individual XML files.

## Description

This tool reads a `db.xml` file (typically found in Fantasy Grounds campaign folders), parses the `<charsheet>` section, and extracts each character into its own file named `character_<ID>_<Level>.xml`.

It filters out sensitive or unnecessary tags like `<public>` and `<holder>` information during extraction.

## Setup & Usage

1.  **Build the tool:**
    ```bash
    go build -o fg-char-extract main.go
    ```

2.  **Run extraction:**
    Place your `db.xml` file in the same directory as the executable (or run from the directory containing `db.xml`).
    ```bash
    ./fg-char-extract
    ```

3.  **Output:**
    The tool will generate XML files in the current directory, for example:
    - `character_id-00001_5.xml`
    - `character_id-00002_3.xml`

## Development

- **Language:** Go 1.22+
- **Dependencies:** Standard library only (`encoding/xml`)

## License

MIT
