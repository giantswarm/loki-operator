package project

var (
	bundleVersion = "0.0.1"
	description   = "The loki-operator does something."
	gitSHA        = "n/a"
	name          = "loki-operator"
	source        = "https://github.com/giantswarm/loki-operator"
	version       = "n/a"
)

func BundleVersion() string {
	return bundleVersion
}

func Description() string {
	return description
}

func GitSHA() string {
	return gitSHA
}

func Name() string {
	return name
}

func Source() string {
	return source
}

func Version() string {
	return version
}
