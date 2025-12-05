# aoccli

Small TUI to view an Advent of Code private leaderboard.

## Install (with Go)

Requires Go 1.24+ (or `gotip`) because of upstream deps. Then:

```sh
go install github.com/LakshyaMittal3301/aoccli/cmd/aoccli@latest
```

The binary is written to `$GOBIN` (or `$GOPATH/bin`). Ensure that directory is on your `PATH` so you can run `aoccli` from anywhere.

## Install (prebuilt binaries)

1) Download the archive for your OS/arch from the GitHub Releases page (`aoccli_<tag>_<os>_<arch>.tar.gz` or `.zip`).
2) Extract it to get the `aoccli` (or `aoccli.exe`) binary.
3) (macOS/Linux) `chmod +x aoccli` and move it somewhere on `PATH`, e.g. `mv aoccli /usr/local/bin` or `~/.local/bin`.
4) (Windows) Keep `aoccli.exe` in a folder on `PATH` or run it from its directory.

## Usage

Run `aoccli`. On first launch you’ll be prompted for your private leaderboard JSON URL (AoC share link). The URL is saved at `~/.config/aoccli/config.json` (or `$XDG_CONFIG_HOME/aoccli/config.json`).

Controls: `q` to quit, `←/→` to change days, `d` to open the day list, `r` to refresh.

To update, reinstall via `go install ...@latest` or replace the binary with a newer release download.

### CLI flags

- `-h`/`--help`: show usage and exit.
- `-reset-config`: delete the saved config file and exit (useful if you need to re-enter the URL).
