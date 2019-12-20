package flag

import (
	"github.com/giantswarm/loki-operator/flag/loki"
	"github.com/giantswarm/loki-operator/flag/service"
	"github.com/giantswarm/microkit/flag"
)

// Flag provides data structure for service command line flags.
type Flag struct {
	Service service.Service
	Loki    loki.Loki
}

// New constructs fills new Flag structure with given command line flags.
func New() *Flag {
	f := &Flag{}
	flag.Init(f)

	return f
}
