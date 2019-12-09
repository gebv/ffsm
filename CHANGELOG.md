# Changelog
All notable changes to this project will be documented in this file.

## [Unreleased]

- added debug stack for `dispatcherError` (for debug if panic in handler)

## [1.2.1] - 2019-12-08

### Added
- added simple FSM and separate from FSM with handlers
- updated README with example codes
- more tests

### Changed
- moved Dispatch to AsyncDispatch and Dispatch for sync requests

### Removed
- removed the ability to call the dispatcher in the handler
- removed check for transition to oneself and error ErrTransitionToItSelf
- removed package github.com/pkg/errors
