# aoccli

Small TUI to view an Advent of Code private leaderboard.

## Install (with Go)

Requires Go 1.24+ (or `gotip`) because of upstream deps. Install from the `main` branch (we are not publishing release binaries right now):

```sh
go install github.com/LakshyaMittal3301/aoccli/cmd/aoccli@main
```

The binary is written to `$GOBIN` (or `$GOPATH/bin`). Ensure that directory is on your `PATH` so you can run `aoccli` from anywhere.

## Install (prebuilt binaries)

Prebuilt release archives are currently disabled; prefer `go install ...@main`. If/when releases are re-enabled, download the OS/arch archive from the Releases page, extract the binary, and place it on your `PATH`.

## Usage

Run `aoccli`. On first launch you’ll be prompted for your private leaderboard JSON URL (AoC share link). The URL is saved at `~/.config/aoccli/config.json` (or `$XDG_CONFIG_HOME/aoccli/config.json`).

Controls: `q` to quit, `←/→` to change days, `d` to open the day list, `r` to refresh.

To update, reinstall via `go install ...@main`.

### CLI flags

- `-h`/`--help`: show usage and exit.
- `-reset-config`: delete the saved config file and exit (useful if you need to re-enter the URL).
