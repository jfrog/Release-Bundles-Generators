# release-bundle-generator

This project is designed to simplify interaction with release bundles, by
generating them from other formats. Currently, it can generate release bundles
from Helm charts.

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
   - On Unix based systems, run:
     ```
     cd release-bundle-generators
     ./build.sh
     ```
   - On Windows, run:
     ```
     cd release-bundle-generators
     build.bat
     ```

## Usage

The release-bundle-generator can be configured in much the same way as
[JFrog CLI](https://jfrog.com/getcli): you can configure an Artifactory instance
with:

``` shell
release-bundle-generator c
```

and you can switch between configured instances with:

``` shell
release-bundle-generator use <server ID>
```

Alternatively, you can use the `--server-id=<server ID>` option to pass the
configuration in directly, without setting it as default. Or, you can use the
other optional arguments to pass options in without a configuration.

To generate a release bundle from a chart, you can run:

``` shell
release-bundle-generator from-chart --chart-path=<chart path> --docker-repo=<Docker repo name> <bundle name> <bundle version>
```

This has the following parameters:
- The path (in Artifactory) of the Helm chart to generate. All dependency Helm
  charts should be available in the same repository.
- The name of a Docker repository in Artifactory. All dependency Docker images
  should be available in this repository.
- The name and version that the new release bundle should have.
- All optional arguments can be checked by running `release-bundle-generator from-chart --help`

Note that if Helm or Docker dependencies are found in a remote repository, they
must be cached. Otherwise, they won't show up in the release bundle. After
generating a release bundle, the generator will output which dependencies were
and were not found; missing dependencies are not listed in the bundle.
