# Install (Detailed)

This file is intentionally separate from `SKILLS.md` so one-time setup does not clutter the agent quickstart.

## Homebrew (macOS)

```bash
brew tap rintays/tap
brew install rintays/tap/lark
```

## Build from source

```bash
git clone https://github.com/rintays/lark-cli.git
cd lark
go install ./cmd/lark
lark --help
```

Local binary:

```bash
go build -o lark ./cmd/lark
./lark --help
```

## GitHub Releases

Download the archive for your OS from the releases page, extract it, and move `lark` into your PATH.

## Verify

```bash
lark --help
lark auth --help
```
