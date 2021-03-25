# Contributing

Issues, whether bugs, tasks, or feature requests are essential for keeping rbac-manager great. We believe it should be as easy as possible to contribute changes that get things working in your environment. There are a few guidelines that we need contributors to follow so that we can keep on top of things.

## Code of Conduct

This project adheres to a [code of conduct](/contributing/code-of-conduct). Please review this document before contributing to this project.

## Sign the CLA
Before you can contribute, you will need to sign the [Contributor License Agreement](https://cla-assistant.io/fairwindsops/rbac-manager).

## Project Structure

rbac-manager is a relatively simple cobra cli tool that looks up information about rbac in a cluster. The [/cmd](https://github.com/FairwindsOps/rbac-manager/tree/master/cmd/manager) folder contains the flags and other cobra config.

## Getting Started

We label issues with the ["good first issue" tag](https://github.com/FairwindsOps/rbac-manager/labels/good%20first%20issue) if we believe they'll be a good starting point for new contributors. If you're interested in working on an issue, please start a conversation on that issue, and we can help answer any questions as they come up.

## Setting Up Your Development Environment
### Prerequisites
* A properly configured Golang environment with Go 1.11 or higher
* Access to a cluster via a properly configured KUBECONFIG

### Installation
* Install the project with `go get github.com/fairwindsops/rbac-manager`
* Change into the rbac-manager directory which is installed at `$GOPATH/src/github.com/fairwindsops/rbac-manager`
* Run tests with `make test`

## Creating a New Issue

If you've encountered an issue that is not already reported, please create an issue that contains the following:

- Clear description of the issue
- Steps to reproduce it
- Appropriate labels

## Creating a Pull Request

Each new pull request should:

- Reference any related issues
- Add tests that show the issues have been solved
- Pass existing tests and linting
- Contain a clear indication of if they're ready for review or a work in progress
- Be up to date and/or rebased on the master branch

## Creating a new release

Push a new annotated tag.  This tag should contain a changelog of pertinent changes. Currently github releases are manual; go into the UI and create a release with the changelog as the text. In the future, it would be nice to implement Goreleaser.
