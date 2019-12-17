package promtailconfig

import "time"

// Key allows to identify a promtail config for a specific ContainerName running
// in a pod selectable by Labels in Namespace.
// Key's values must be rendered as comments into the final config map
// so it's possible to load the map from a config file and recreate existing keys.
// To make the Keys easily comparable and possible to use as map keys,
// Labels are not stored as "map[string]string", but just string of format
// "k1=v1,k2=v2,..."
type Key struct {
	Namespace     string
	Labels        string
	ContainerName string
}

// Handler is an interface that delivers operations required to sync between
// events created by pods with related configmap and the actual promtail's
// configmap.
type Handler interface {
	AddConfig(key Key, yamlContent string)
	DelConfig(key Key, yamlContent string)
	Render() string
}

// PeriodicHandler is an implementation of handler, that loads promtail's configmap
// at the beginning, parses it and stores with appropriate Keys in the snippet field.
// Then it allows for AddConfig/DelConfig calls to be made.
// PeriodicHandler has a configurable timer, that periodically renders all the
// data in snippets and produces a new value for the promtail's configmap.
type PeriodicHandler struct {
	snippets map[Key]string
}

func NewPeriodicHandler(initialDelay time.Duration, period time.Duration) (*Handler, error) {
	return nil, nil
}
