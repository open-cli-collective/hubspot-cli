# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed
- `--output json` now writes status/progress messages to stderr instead of stdout, so stdout is valid, parseable JSON (#52)
- `associations list` no longer fails to parse responses with results: `toObjectId` is now decoded as a number (#51)

### Changed
- Improved init/config UX with huh forms, config pre-population, and --force flag on clear (#44)
- Removed `config set` command - use `init` for configuration changes (#44)
- Updated token masking format for consistency (#44)

### Documentation
- Added comprehensive README.md (#42)

### Internal
- Code simplification and consistency improvements (#40)
