# Release

This project uses GoReleaser and GitHub Actions to publish tagged releases.

## Tag-based releases

1. Create a tag in the form `vX.Y.Z` on the commit you want to release.
2. Push the tag to GitHub.
3. The `release` workflow runs GoReleaser, builds cross-platform binaries, and publishes a GitHub Release with checksums.

Example:

```bash
git tag v1.2.3
git push origin v1.2.3
```

## Manual re-release

The `release` workflow also supports manual runs. Trigger it via GitHub Actions and provide a tag like `v1.2.3`.
The workflow checks out the tag and uses the GoReleaser config from the default branch.

## Version value

`lark version` prints the value injected by GoReleaser via:

- `-X lark/internal/version.Version={{ .Tag }}`
- `-X lark/internal/version.Commit={{ .ShortCommit }}`
- `-X lark/internal/version.Date={{ .Date }}`
