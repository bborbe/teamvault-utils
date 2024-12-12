# Changelog

All notable changes to this project will be documented in this file.

Please choose versions by [Semantic Versioning](http://semver.org/).

* MAJOR version when you make incompatible API changes,
* MINOR version when you add functionality in a backwards-compatible manner, and
* PATCH version when you make backwards-compatible bug fixes.

## v1.9.1

- add Time.Sub 
- add Duration.Abs

## v1.9.0

- Duration.String output now days and weeks like: 10w5d23h59m30s
- go mod update

## v1.8.2

- add list types

## v1.8.1

- allow parse timeOfDay with seconds

## v1.8.0

- allow unmarshal NOW
- go mod update

## v1.7.4

- add parse time without seconds
- go mod update

## v1.7.3

- Add MarshalBinary and UnmarshalBinary
- go mod update

## v1.7.2

- add Date.Add() and UnixTime.Add()

## v1.7.1

- add Year(), Month(), Day(), Hour(), Minute(), Second() and Nanosecond()

## v1.7.0

- add ParseUnixTime
- go mod update

## v1.6.2

- add Before, After and Equal to TimeOfDay

## v1.6.1

- add Duration.String()

## v1.6.0

- add ParseDateTimeDefault
- add ParseDurationDefault
- add ParseTimeDefault
- add ParseTimeOfDayDefault

## v1.5.2

- remove error from DateTime
- add Time

## v1.5.1

- add DateTime to TimeOfDay

## v1.5.0

- add Duration
- marshal unmarshal Duration from duration string or number
- go mod update

## v1.4.2

- test marshal Date and DateTime

## v1.4.1

- fix parse empty Date and DateTime

## v1.4.0

- add DateTime
- add UnixTime
- test Date
- go mod update

## v1.3.0

- Add compare time
- go mod update

## v1.2.0

- Allow ParseTimeOfDay with Timezone. Example '13:37 Europe/Berlin'
- go mod update

## v1.1.1

- fix ParseDuration

## v1.1.0

- Add ParseDuration with support for d=day and w=week

## v1.0.0

- Initial Version
