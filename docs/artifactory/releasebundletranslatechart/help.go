package releasebundletranslatechart

const Description = "Generate a release bundle from an existing Helm chart."

var Usage = []string{"release-bundle-generator fc --chart-path=<Path to Helm chart> --docker-repo=<Docker repo name> [command options] <release bundle name> <release bundle version>"}

const Arguments string = `	release bundle name
		The name of the release bundle.

	release bundle version
		The release bundle version.`
