# Changelog

All notable changes to this project will be documented in this file.

Please choose versions by [Semantic Versioning](http://semver.org/).

* MAJOR version when you make incompatible API changes,
* MINOR version when you add functionality in a backwards-compatible manner, and
* PATCH version when you make backwards-compatible bug fixes.

## v1.12.2

- go mod update

## v1.12.1

- go mod update
- add test for relation store mocks

## v1.12.0

- add Invert for the RelationStore and RelationStoreTx

## v1.11.5

- add MapIDRelations and MapRelationIDs

## v1.11.4

- remove performance bug in relationStoreTx delete

## v1.11.3

- move JsonHandlerTx to github.com/bborbe/http
- go mod update

## v1.11.2

- add missing license file
- go mod update

## v1.11.1

- rename NewUpdateHandlerViewTx -> NewJsonHandlerUpdateTx

## v1.11.0

- add JsonHandlerTx
- go mod update

## v1.10.0

- add ListBucketNames
- go mod update

## v1.9.1

- ignore BucketNotFoundError on Map, Remove and Exists
- go mod update

## v1.9.0

- add remove to DB to delete the complete database
- add handler for reset bucket and complete database
- go mod update

## v1.8.2

- fix replace in relationStore
- go mod update

## v1.8.1

- add simple benchmark

## v1.8.0

- add relation store
- go mod update

## v1.7.0

- expect same tx returns same bucket
- go mod update

## v1.6.0

- add stream and exists to store

## v1.5.0

- add JSON store
- go mod update

## v1.4.2

- add KeyNotFoundError

## v1.4.1

- add mocks

## v1.4.0

- expect error if transaction open second transaction

## v1.3.1

- improve iterator testsuite

## v1.3.0

- add bucket testsuite

## v1.2.0

- add provider
- improve testsuite

## v1.1.1

- add test for iterator seek not found

## v1.1.0

- Add context to update and view

## v1.0.0

- Initial Version
