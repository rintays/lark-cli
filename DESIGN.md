# Design

This CLI follows a gog-inspired layout: a single root command wires shared state and subcommands, and the core behavior lives under internal packages.

## Structure

- cmd/lark: Cobra commands and wiring. The root command builds an appState and shared flags. Subcommands live beside it.
- internal/config: Load/save config (default path, JSON format, secure permissions).
- internal/larkapi: Minimal HTTP client for Lark APIs (tenant token, whoami, message send).
- internal/output: Unified output handling (human text vs JSON).
- internal/testutil: Test HTTP helpers.

## Runtime flow

1. Root command resolves the config path, loads config, and builds the API client.
2. Commands call helper functions that enforce credentials and reuse cached tokens when valid.
3. Token refresh updates config on disk.

## Config + caching

- Default config path: ~/.config/lark/config.json
- Token caching is stored in config and reused until close to expiry.
- Commands are written to accept an alternate config path via the global --config flag.
