# DEPRECATION NOTICE

This repo has been deprecated as the release bundle generator as migrated to a jfrog cli plugin.

You can find the new repo at [JFrog CLI Plugins](https://github.com/jfrog/jfrog-cli-plugins/tree/main/rb-gen)

# release-bundle-generator

This project is designed to simplify interaction with release bundles, by
generating them from other formats. Currently, it can generate release bundles
from Helm charts.

This project is a [JFrog CLI](https://github.com/jfrog/jfrog-cli) plugin, and
can be installed and run from within the JFrog CLI, or as a standalone
executable:

## Installation with JFrog CLI
Installing the latest version:

`$ jfrog plugin-install release-bundle-generator`

Installing a specific version:

`$ jfrog plugin-install release-bundle-generator@version`

Uninstalling a plugin

`$ jfrog plugin-uninstall release-bundle-generator`

## Downloading the Executable

MacOS and Linux builds are available from Bintray,
[here](https://bintray.com/darthfennec/jfrog-release-bundle-generator/release-bundle-generator).

## Building the Executable

1. [Install Go](https://golang.org/doc/install).
2. Verify that the `GO111MODULE` variable is either unset, or explicitly set to
   `auto`.
3. Clone the `release-bundle-generators` project by executing the following command:
   ```
   git clone https://github.com/jfrog/release-bundle-generators
   ```
4. Build the project by navigating to the `release-bundle-generators` folder and
   executing the build script:
   ```
   cd release-bundle-generators
   go build
   ```

## Usage

To generate a release bundle from a chart, you can run:

``` shell
jfrog from-chart --chart-path=<chart path> --docker-repo=<Docker repo name> <bundle name> <bundle version>
```

Or, if you're running it as a standalone executable:

``` shell
./release-bundle-generator from-chart --chart-path=<chart path> --docker-repo=<Docker repo name> <bundle name> <bundle version>
```

This has the following parameters:
- The path (in Artifactory) of the Helm chart from which to generate the release
  bundle. All dependency Helm charts should be available in the same repository.
- The name of a Docker repository in Artifactory. All dependency Docker images
  should be available in this repository.
- The name and version that the new release bundle should have.
- All optional arguments can be checked by running `jfrog from-chart --help`

Note that if Helm or Docker dependencies are found in a remote repository, they
must be cached. Otherwise, they won't show up in the release bundle. After
generating a release bundle, the generator will output which dependencies were
and were not found; missing dependencies are not listed in the bundle.
