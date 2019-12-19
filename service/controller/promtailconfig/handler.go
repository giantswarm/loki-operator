package promtailconfig

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/giantswarm/microerror"
	v1 "k8s.io/api/core/v1"
)

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

func NewKey(pod *v1.Pod, containerName string) *Key {
	var labels strings.Builder
	labelKeys := []string{}
	for k := range pod.ObjectMeta.Labels {
		labelKeys = append(labelKeys, k)
	}
	sort.Strings(labelKeys)
	for _, k := range labelKeys {
		v := pod.ObjectMeta.Labels[k]
		labels.WriteString(fmt.Sprintf("%v=%v,", k, v))
	}

	return &Key{
		Namespace:     pod.Namespace,
		Labels:        labels.String(),
		ContainerName: containerName,
	}
}

func SortKeys(keys []Key) {
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Namespace < keys[j].Namespace {
			return true
		} else if keys[i].Namespace > keys[j].Namespace {
			return false
		}
		// here it means Namespace values are equal, move to the next one
		if keys[i].Labels < keys[j].Labels {
			return true
		} else if keys[i].Labels > keys[j].Labels {
			return false
		}
		// here it means Labels values are equal, move to the next one
		return keys[i].ContainerName < keys[j].ContainerName
	})
}

// Handler is an interface that delivers operations required to sync between
// events created by pods with related configmap and the actual promtail's
// configmap.
type Handler interface {
	AddConfig(key Key, yamlContent string)
	DelConfig(key Key)
}

// PeriodicHandler is an implementation of handler, that loads promtail's configmap
// at the beginning, parses it and stores with appropriate Keys in the snippet field.
// Then it allows for AddConfig/DelConfig calls to be made.
// PeriodicHandler has a configurable timer, that periodically renders all the
// data in snippets and produces a new value for the promtail's configmap.
type PeriodicHandler struct {
	snippets     map[Key]string
	initialDelay time.Duration
	period       time.Duration
	promMap      *PromtailConfigMap
	updateTimer  *time.Timer
}

func NewPeriodicHandler(initialDelay time.Duration, period time.Duration, promMap *PromtailConfigMap) (Handler, error) {
	if initialDelay <= 0 {
		return nil, microerror.New("initialDelay must be > 0")
	}
	if period <= 0 {
		return nil, microerror.New("period must be > 0")
	}
	if promMap == nil {
		return nil, microerror.New("promMap can't be nil")
	}

	ph := &PeriodicHandler{
		snippets:     make(map[Key]string),
		initialDelay: initialDelay,
		period:       period,
		promMap:      promMap,
	}
	if err := ph.init(); err != nil {
		return nil, err
	}
	return ph, nil
}

func (p *PeriodicHandler) AddConfig(key Key, yamlContent string) {
	p.snippets[key] = yamlContent
}

func (p *PeriodicHandler) DelConfig(key Key) {
	if _, found := p.snippets[key]; found {
		delete(p.snippets, key)
	}
}

func (p *PeriodicHandler) init() error {
	snips, err := p.promMap.Load()
	if err != nil {
		return err
	}
	p.snippets = snips

	time.AfterFunc(p.initialDelay, func() {
		p.updateTimer = time.NewTimer(p.period)
		p.promMap.Update(p.snippets)
		go p.handleUpdateTimer()
	})
	return nil
}

func (p *PeriodicHandler) handleUpdateTimer() {
	for {
		<-p.updateTimer.C
		p.promMap.Update(p.snippets)
	}
}
