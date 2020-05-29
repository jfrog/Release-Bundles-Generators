package common

const GlobalEnvVars string = `	JFROG_BUNDLE_GEN_HOME_DIR
		[Default: ~/.jfrog/bundlegen]
		Defines the the release bundle generator home directory path.

	JFROG_BUNDLE_GEN_TEMP_DIR
		[Default: The operating system's temp directory]
		Defines the temp directory used by the release bundle generator.

	CI
		[Default: false]
		If true, disables interactive prompts and progress bar.
		`
