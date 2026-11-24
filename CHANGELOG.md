# Changelog

All notable changes to this project are documented here. This changelog follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project uses
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Add Makefile targets to build Docker images locally and publish versioned and
  `latest` images to Docker Hub and GHCR.

## [0.1.1] - 2026-11-24

### Added

- Add GitHub issue and pull request templates, code ownership rules, and an
  automated dependency update workflow.
- Publish multi-platform Docker images for `linux/amd64` and `linux/arm64` to
  Docker Hub and GHCR.

### Changed

- Use the Distroless Debian 12 nonroot image for the Docker runtime.

## [0.1.0] - 2026-06-26

### Added

- Initial PocketID to Listmonk sync sidecar.
- Reconcile subscribers by creating, updating, and deleting Listmonk entries
  based on PocketID users.
- Add dry-run mode and environment variable configuration.
- Add Docker, Kubernetes CronJob, Docker Compose, and cron usage documentation.
