# Changelog

All notable changes to this project will be documented in this file.

Please choose versions by [Semantic Versioning](http://semver.org/).

* MAJOR version when you make incompatible API changes,
* MINOR version when you add functionality in a backwards-compatible manner, and
* PATCH version when you make backwards-compatible bug fixes.

## v4.6.2

- Add `make all` target to run precommit checks and install binaries
- Reorganize Makefile structure
- Update dependencies

## v4.6.1

- Move NormalizePath function into package (remove external dependency)
- Remove dependency on github.com/bborbe/io and github.com/bborbe/assert
- Update Go version to 1.25.2

## v4.6.0

- Add GitHub workflows for CI, Claude Code review, and Claude
- Add golangci-lint configuration
- Add key validation with context support
- Add gosec suppressions for controlled file reads
- Update dependencies
- Update Makefile with security checks
- Update all commands to use libservice.MainCmd
- Add copyright headers to all files

## v4.5.3

- use service.MainCmd

## v4.5.2

- remove sentry
- prevent print args

## v4.5.1

- fix teamvault-config-parser

## v4.5.0

- go mod update
- use lib argument

## v4.4.4

- go mod update

## v4.4.3

- go mod update

## v4.4.2

- go mod update

## v4.4.1

- go mod update

## v4.4.0

- fix go module to github.com/bborbe/teamvault-utils/v4 

## v4.3.3

- go mod update
- remove deprecated golint

## v4.3.2

- refactor

## v4.3.1

- go mod update

## v4.3.0

- go mod update
- inline lib http helper
- refactor

## v4.2.0

- add cache option for secrets

## v4.1.1

- update all deps

## v4.1.0

- update all deps
- go version to 1.21

## v4.0.1

- update all deps
- go version to 1.19

## v4.0.0

- add teamvault-file command
- remove subpackages
- use go modules instead dep

## v3.4.0

- add readfile to read content from file
- add indent method

## v3.3.0

- Add Htpasswd generator 

## v3.2.0

- Add cache connector

## v3.1.1

- Create fallback dirs

## v3.1.0

- Add disk fallback connector

## v3.0.1

- Update deps

## v3.0.0

- Move mode and Connector interface to root

## v2.1.0

- add search method to connector

## v2.0.0

- rename bin to cmd
- replace unterscore with dash in commands
- check config file is no directory 

## v1.2.1

- fix commands

## v1.2.0

- add teamvault_username, teamvault_password and teamvault_url command

## v1.1.0

- Add teamvaultHtpasswd

## v1.0.0

- Initial version
