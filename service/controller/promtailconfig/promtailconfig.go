package promtailconfig

import (
	"fmt"
	"strings"

	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	nsHeader        = "# loki-operator.namespace"
	containerHeader = "# loki-operator.container"
	labelsHeader    = "# loki-operator.labels"
	configMapHeader = `# this config is auto-generated by loki-operator - manual changes WILL BE LOST
client:
  backoff_config:
    maxbackoff: 5s
    maxretries: 20
    minbackoff: 100ms
  batchsize: 102400
  batchwait: 1s
  external_labels: {}
  timeout: 10s
positions:
  filename: /run/promtail/positions.yaml
server:
  http_listen_port: 3101
target_config:
  sync_period: 10s
scrape_configs:
`
)

type PromtailConfigMap struct {
	k8sClient     k8sclient.Interface
	namespace     string
	name          string
	configKeyName string
}

func NewPromtailConfigMap(k8sClient k8sclient.Interface, namespace, name, configKeyName string) (*PromtailConfigMap, error) {
	if name == "" {
		return nil, &microerror.Error{
			Desc: "Name of the promtail config map can't be empty",
		}
	}
	if namespace == "" {
		return nil, &microerror.Error{
			Desc: "Namespace of the promtail config map can't be empty",
		}
	}
	if configKeyName == "" {
		return nil, &microerror.Error{
			Desc: "Name of the configmap key in promtail config map can't be empty",
		}
	}
	if k8sClient == nil {
		return nil, &microerror.Error{
			Desc: "k8sClient can't be nil",
		}
	}
	return &PromtailConfigMap{
		k8sClient:     k8sClient,
		namespace:     namespace,
		name:          name,
		configKeyName: configKeyName,
	}, nil
}

func (p *PromtailConfigMap) Load() (map[Key]string, error) {
	cm, err := p.loadConfigMap()
	if err != nil {
		return nil, err
	}
	config, found := cm.Data[p.configKeyName]
	if !found {
		// TODO: log here
		return nil, nil
	}
	var lines []string
	for _, l := range strings.Split(config, "\n") {
		if l == "" {
			continue
		}
		lines = append(lines, l)
	}

	res := make(map[Key]string)
	startLineIndex := -1
	for startLineIndex = range lines {
		if lines[startLineIndex] == "scrape_configs:" {
			startLineIndex++
			break
		}
	}
	for startLineIndex < len(lines) {
		key, err := p.parseKey(lines, startLineIndex)
		if err != nil {
			return nil, err
		}
		startLineIndex += 3
		var nextStart int
		for nextStart = startLineIndex;;nextStart++ {
			if strings.HasPrefix(lines[nextStart], containerHeader) {
				break
			}
		}
		cfg := strings.Join(lines[startLineIndex:nextStart], "\n")
		res[key] = cfg
		startLineIndex = nextStart
	}
	return res, nil
}

func (p *PromtailConfigMap) parseKey(lines []string, startIndex int) (Key, error) {
	if !(strings.HasPrefix(lines[startIndex], containerHeader) &&
		strings.HasPrefix(lines[startIndex+1], nsHeader) &&
		strings.HasPrefix(lines[startIndex+2], labelsHeader)) {
		return Key{}, microerror.New("Couldn't find expected header")
	}
	key := Key{
		ContainerName: lines[startIndex][len(containerHeader)+1:],
		Namespace:     lines[startIndex+1][len(nsHeader)+1:],
		Labels:        lines[startIndex+2][len(labelsHeader)+1:],
	}
	return key, nil
}

func (p *PromtailConfigMap) Update(newSnippets map[Key]string) error {
	oldSnippets, err := p.Load()
	if err != nil {
		return err
	}

	updateNeeded := false
	if len(oldSnippets) != len(newSnippets) {
		updateNeeded = true
	} else {
		var oldKeysSorted, newKeysSorted []Key
		for k, _ := range oldSnippets {
			oldKeysSorted = append(oldKeysSorted, k)
		}
		for k, _ := range newSnippets {
			newKeysSorted = append(newKeysSorted, k)
		}
		SortKeys(oldKeysSorted)
		SortKeys(newKeysSorted)
		for i := range newKeysSorted {
			// if relevant keys or values for these keys differ, we need to update
			if oldKeysSorted[i] != newKeysSorted[i] || oldSnippets[oldKeysSorted[i]] != newSnippets[newKeysSorted[i]] {
				updateNeeded = true
				break
			}
		}
	}

	if !updateNeeded {
		return nil
	}
	return p.save(newSnippets)
}

func (p *PromtailConfigMap) loadConfigMap() (*v1.ConfigMap, error) {
	cm, err := p.k8sClient.K8sClient().CoreV1().ConfigMaps(p.namespace).Get(p.name, metav1.GetOptions{})
	if err != nil {
		return nil, microerror.Maskf(err, "Couldn't load configmap %s/%s", p.namespace, p.name)
	}
	return cm, nil
}

func (p *PromtailConfigMap) render(key Key, snippet string) string {
	var config strings.Builder
	config.WriteString(fmt.Sprintf("%s %s\n", containerHeader, key.ContainerName))
	config.WriteString(fmt.Sprintf("%s %s\n", nsHeader, key.Namespace))
	config.WriteString(fmt.Sprintf("%s %s\n", labelsHeader, key.Labels))
	lines := strings.Split(snippet, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		config.WriteString(fmt.Sprintf("%s\n", snippet))
	}
	config.WriteString("\n")

	return config.String()
}

func (p *PromtailConfigMap) save(snippets map[Key]string) error {
	var config strings.Builder
	config.WriteString(configMapHeader)
	for key, snippet := range snippets {
		config.WriteString(p.render(key, snippet))
	}

	cm, err := p.loadConfigMap()
	if err != nil {
		return err
	}

	cm.Data[p.configKeyName] = config.String()
	if _, err := p.k8sClient.K8sClient().CoreV1().ConfigMaps(p.namespace).Update(cm); err != nil {
		return microerror.Maskf(err, "Couldn't update promtail configmap %s/%s", p.namespace, p.name)
	}

	return nil
}
