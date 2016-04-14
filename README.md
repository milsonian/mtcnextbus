# mtcnextbus

A simple go program that returns remaining time to next departing MTC bus or train, given a route, stop, and direction

## Installation

1. (Prerequisite) Install Go and configure GOPATH - see https://golang.org/dl/
2. `go get github.com/milsonian/mtcnextbus` (or `go install` from project path if cloned)

Alternatively, you can use Docker (Go development environment not required)
1. (Prerequisite) Install Docker - see https://docs.docker.com/engine/installation/
2. `docker pull milsonian/mtcnextbus` (optional, image will also pull at first run)

## Usage

mtcnextbus -route "*route name*" -stop "*stop name*" -direction "*cardinal direction*"

Example:  `mtcnextbus -route "METRO Blue Line" -stop "Target Field Station Platform 1" -direction "north"`

This returns e.g. `next departure: 34 Min (per schedule)` or an error if there are no remaining departures for that day.  `(Actual vehicle report)` indicates a real-time vehicle-GPS reported result.

Note: This app depends on internet connectivity to the MTC NexTrip API to function (http://svc.metrotransit.org/)

## Usage (Docker)

docker run -it milsonian/mtcnextbus -route "*route name*" -stop "*stop name*" -direction "*cardinal direction*"

Example: `docker run milsonian/mtcnextbus -route "METRO Blue Line" -stop "Target Field Station Platform 1" -direction "north"`

## Development & Contributing
1. Install Go and configure GOPATH if you haven't - (step 1 of Installation above)
2. Fork this Repo, then clone: `git clone https://github.com/user/pathtorepo.git`
3. Create your feature branch: `git checkout -b my-new-feature`
4. Commit your changes, including necessary tests: `git commit -am 'Add some feature'`
5. Push to the branch: `git push origin my-new-feature`
6. Submit a pull request - please include result of test run

## Testing
`go test` in your project path or `go test github.com/user/mtcnextbus`

## TODO

* Version and add changelog
* Improve/consolidate error handling
* Add debug logging
* Increase test coverage, mock API for tests
* Restructure/reduce duplication (in e.g. unmarshalling, result handling)
* Cache certain responses/reduce spam to backing API
* Return valid routes/stops with flag (similar to behavior with directions)
* Return provider string along with departure (plumbed)
* Add http interface
