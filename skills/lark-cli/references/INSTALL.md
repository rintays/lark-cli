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

macOS/Linux example:

```bash
curl -L https://github.com/rintays/lark-cli/releases/latest/download/lark_<VERSION>_darwin_arm64.tar.gz -o lark.tar.gz
tar -xzf lark.tar.gz
chmod +x lark
sudo mv lark /usr/local/bin/lark
```

Windows (PowerShell) example:

```powershell
Invoke-WebRequest -Uri https://github.com/rintays/lark-cli/releases/latest/download/lark_<VERSION>_windows_amd64.zip -OutFile lark.zip
Expand-Archive lark.zip -DestinationPath .
Move-Item .\lark.exe $env:USERPROFILE\bin\lark.exe
```

## Verify

```bash
lark --help
lark auth --help
```
