package test

import (
	"github.com/giantswarm/microerror"
)

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var invalidDynamicConfigError = &microerror.Error{
	Kind: "invalidDynamicConfigError",
}

// IsInvalidDynamicConfig asserts invalidConfigError.
func IsInvalidDynamicConfig(err error) bool {
	return microerror.Cause(err) == invalidDynamicConfigError
}
