# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Mock implementations for unit testing (`mocks/mocks.go`)
- `Health()` method to DI container for service health checks
- GitHub Actions CI workflow with lint, test, and coverage
- URL validation tests with SSRF protection coverage
- Rate limiting for Playwright crawler

### Changed

- Updated Playwright client to use locator-based APIs (replaced deprecated page-level methods)
- Fixed spider.go to avoid embedded field access pattern

### Added

- CONTRIBUTING.md with contribution guidelines
- CHANGELOG.md for version history tracking
- URL validation for crawler security

## [1.0.0] - 2025-12-27

### Added

- Initial release
- Dependency injection container with conditional initialization
- Cache implementations: LRU, Redis
- Database clients: MySQL, PostgreSQL, ClickHouse, BigTable
- Message queue clients: Kafka, RabbitMQ
- Crawler implementations: Colly, Playwright, Selenium, Puppeteer, Spider, Soup, Ferret
- Configuration management with Viper
- Structured logging with Zap
- Temporal workflow integration
- Comprehensive test suite
- Docker Compose setup for development
