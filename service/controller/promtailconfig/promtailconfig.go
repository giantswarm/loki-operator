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
	configMapHeader = `
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
	panic("Not implemented!")
}

func (p *PromtailConfigMap) Update(currentSnippets map[Key]string) error {
	//TODO: load the config map an check if updated is needed
	updateNeeded := true
	if !updateNeeded {
		return nil
	}
	return p.save(currentSnippets)
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
	config.WriteString(fmt.Sprintf("# %s\n", key.ContainerName))
	config.WriteString(fmt.Sprintf("# %s\n", key.Namespace))
	config.WriteString(fmt.Sprintf("# %s\n", key.Labels))
	config.WriteString(snippet)
	config.WriteRune('\n')

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
