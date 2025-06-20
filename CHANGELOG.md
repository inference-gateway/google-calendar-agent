# Changelog

All notable changes to this project will be documented in this file.

## [0.3.0-rc.3](https://github.com/inference-gateway/google-calendar-agent/compare/v0.3.0-rc.2...v0.3.0-rc.3) (2025-06-20)

### 🐛 Bug Fixes

* Update a2a dependency to v0.4.0-rc.10 ([c61de63](https://github.com/inference-gateway/google-calendar-agent/commit/c61de637259fb18e6cb4dbadb4c444986f729050))
* Update a2a dependency to v0.4.0-rc.8 and add streaming task submission example ([941810c](https://github.com/inference-gateway/google-calendar-agent/commit/941810c528d9ad0b8b681c83325637a5784bdf55))
* Update a2a dependency to v0.4.0-rc.9 ([1ba62be](https://github.com/inference-gateway/google-calendar-agent/commit/1ba62be567f6cabea644a83d9d46bd9ed5667200))

### 🔧 Miscellaneous

* Update inference-gateway and google-calendar-agent images to latest version ([3e9d1b5](https://github.com/inference-gateway/google-calendar-agent/commit/3e9d1b56607ef6530129391d3984858cd2f7873f))

## [0.3.0-rc.2](https://github.com/inference-gateway/google-calendar-agent/compare/v0.3.0-rc.1...v0.3.0-rc.2) (2025-06-19)

### 🐛 Bug Fixes

* Update a2a dependency to v0.4.0-rc.3 ([7f52aa3](https://github.com/inference-gateway/google-calendar-agent/commit/7f52aa39219510a162057bff77aa2f9a5f211b51))

## [0.3.0-rc.1](https://github.com/inference-gateway/google-calendar-agent/compare/v0.2.6...v0.3.0-rc.1) (2025-06-19)

### ✨ Features

* Update a2a dependency to v0.4.0-rc.1 ([091934e](https://github.com/inference-gateway/google-calendar-agent/commit/091934e2234d855a5161504f81dd1f8de652e047))

### 📚 Documentation

* Example cleanup ([#10](https://github.com/inference-gateway/google-calendar-agent/issues/10)) ([398ef0d](https://github.com/inference-gateway/google-calendar-agent/commit/398ef0de716468e5fc009df7a778367408fd7031))

## [0.2.6](https://github.com/inference-gateway/google-calendar-agent/compare/v0.2.5...v0.2.6) (2025-06-18)

### 🐛 Bug Fixes

* Refactor agent card retrieval and update dependencies ([fcae906](https://github.com/inference-gateway/google-calendar-agent/commit/fcae9065e7b873d39f073e3cd84ce78f2f8de617))

## [0.2.5](https://github.com/inference-gateway/google-calendar-agent/compare/v0.2.4...v0.2.5) (2025-06-17)

### 🐛 Bug Fixes

* Implement demo mode task handler with mock responses ([e1c9924](https://github.com/inference-gateway/google-calendar-agent/commit/e1c9924b6d16ebcc724e76970fc1ed7c5564bf1a))

## [0.2.4](https://github.com/inference-gateway/google-calendar-agent/compare/v0.2.3...v0.2.4) (2025-06-17)

### 🐛 Bug Fixes

* Correct main path for google-calendar-agent build ([8221f8b](https://github.com/inference-gateway/google-calendar-agent/commit/8221f8bb09760fdb286bb3951daa4d9313dd3177))
* Use agent only if we are not in demo mode ([cd1ec98](https://github.com/inference-gateway/google-calendar-agent/commit/cd1ec98fc90c1cb437ae63274094aef8ff5e4ce4))

## [0.2.3](https://github.com/inference-gateway/google-calendar-agent/compare/v0.2.2...v0.2.3) (2025-06-17)

### ♻️ Improvements

* Split giant file into multiple for clarity ([#9](https://github.com/inference-gateway/google-calendar-agent/issues/9)) ([96ea82d](https://github.com/inference-gateway/google-calendar-agent/commit/96ea82d407dc81466090d50190e7871b12806dc4))

## [0.2.2](https://github.com/inference-gateway/google-calendar-agent/compare/v0.2.1...v0.2.2) (2025-06-17)

### ♻️ Improvements

* Use Inference Gateway A2A ADK ([#6](https://github.com/inference-gateway/google-calendar-agent/issues/6)) ([fee77a8](https://github.com/inference-gateway/google-calendar-agent/commit/fee77a8084466daad960e245045f305399c4878e))

### 📚 Documentation

* Update README with enhanced configuration options and usage examples ([#8](https://github.com/inference-gateway/google-calendar-agent/issues/8)) ([7787deb](https://github.com/inference-gateway/google-calendar-agent/commit/7787deb9a9b2b2dc582b2ceeb62390282f353389))

### 🔧 Miscellaneous

* Remove TODOs related to A2A types and agent template ([6a335b5](https://github.com/inference-gateway/google-calendar-agent/commit/6a335b54ed2d568fee314eee368a40e65750f6e4))
* Update a2a dependency to stable version v0.2.0 ([#7](https://github.com/inference-gateway/google-calendar-agent/issues/7)) ([66cddd4](https://github.com/inference-gateway/google-calendar-agent/commit/66cddd461bec92e1fe75ee317677b620daa8d115))

## [0.2.1](https://github.com/inference-gateway/google-calendar-agent/compare/v0.2.0...v0.2.1) (2025-06-08)

### 📚 Documentation

* Cleanup a bit - remove internal service from direct interaction ([9aba7af](https://github.com/inference-gateway/google-calendar-agent/commit/9aba7af75f662675fa1470461a385d3c4c3ffe63))

### 🔧 Miscellaneous

* Update docker-compose.yaml ([7ffb366](https://github.com/inference-gateway/google-calendar-agent/commit/7ffb36677a5dcc2068dda24df6c3f523779ef578))

### 🔨 Miscellaneous

* Optimize Dockerfile.goreleaser by adding UPX compression for the agent ([#5](https://github.com/inference-gateway/google-calendar-agent/issues/5)) ([e7774fd](https://github.com/inference-gateway/google-calendar-agent/commit/e7774fd646450260913bb971d688fb2e597cfcd5))

## [0.2.0](https://github.com/inference-gateway/google-calendar-agent/compare/v0.1.3...v0.2.0) (2025-06-08)

### ✨ Features

* Enhance CalendarAgent with configuration support and timezone handling ([e7dd96c](https://github.com/inference-gateway/google-calendar-agent/commit/e7dd96c9bfa1b1e994c6e28e2881a6bdb63a4d4a))
* Implement Inference Gateway LLM Service ([#2](https://github.com/inference-gateway/google-calendar-agent/issues/2)) ([04eb229](https://github.com/inference-gateway/google-calendar-agent/commit/04eb22929e68ea2ee3221af771e80497a9c6bd23))

### 🔧 Miscellaneous

* Add timezone configuration for Google Calendar ([64bee3e](https://github.com/inference-gateway/google-calendar-agent/commit/64bee3ed3b1aadf235e36fa293bb5775efd155b9))

## [0.1.3](https://github.com/inference-gateway/google-calendar-agent/compare/v0.1.2...v0.1.3) (2025-06-08)

### ♻️ Improvements

* Update Google credentials handling and configuration options ([c953002](https://github.com/inference-gateway/google-calendar-agent/commit/c953002d318a940e9e421eb55f22292276adc99f))

## [0.1.2](https://github.com/inference-gateway/google-calendar-agent/compare/v0.1.1...v0.1.2) (2025-06-08)

### ♻️ Improvements

* Add configuration tests and utility functions for Google Calendar integration ([504b71c](https://github.com/inference-gateway/google-calendar-agent/commit/504b71c978bcb47d2592e492354e8c61d4f87f06))
* Introduce proper config ([#1](https://github.com/inference-gateway/google-calendar-agent/issues/1)) ([a34a9bd](https://github.com/inference-gateway/google-calendar-agent/commit/a34a9bdedfff096c93530d031e1f8d300b47fbb7))

### 🐛 Bug Fixes

* Correct environment variable name in logging and error messages for credentials file creation ([750c3be](https://github.com/inference-gateway/google-calendar-agent/commit/750c3becbffe10e0995b7f9952573b6096f6e08b))

### 📚 Documentation

* Add early development warning to README ([d597c28](https://github.com/inference-gateway/google-calendar-agent/commit/d597c28671154f2d7a005528b67d8744b62d836f))

### 🔧 Miscellaneous

* Improve cleanup task ([bc68e15](https://github.com/inference-gateway/google-calendar-agent/commit/bc68e15e5abb8b78eb3be44944afcba270c64ca5))

### ✅ Miscellaneous

* Add integration and mock tests for CalendarService functionality ([99b5ad2](https://github.com/inference-gateway/google-calendar-agent/commit/99b5ad23e49b6dfc3cf2fadf741407bad565e910))
* Enhance event request handling and add comprehensive tests for CalendarAgent ([7fcc2b0](https://github.com/inference-gateway/google-calendar-agent/commit/7fcc2b0527e9d0f6c37b791c2549a728f6d033ce))

## [0.1.1](https://github.com/inference-gateway/google-calendar-agent/compare/v0.1.0...v0.1.1) (2025-06-07)

### 🐛 Bug Fixes

* Correct counterfeiter version reference to v6.11.2 in Dockerfile, CI workflow, and goreleaser configuration ([a04daef](https://github.com/inference-gateway/google-calendar-agent/commit/a04daef076bc48dca26a2ef504b28be15890f05c))

### 🔧 Miscellaneous

* Update .gitignore to include /dist and add counterfeiter installation in .goreleaser.yaml ([610bbad](https://github.com/inference-gateway/google-calendar-agent/commit/610bbad23eb8792e039eecab2378af4585310d32))
* Update counterfeiter version to v6.11.2 in Dockerfile, CI workflow, and goreleaser configuration ([c58e832](https://github.com/inference-gateway/google-calendar-agent/commit/c58e832255df3557f1971bc42e0f669bd28653da))
