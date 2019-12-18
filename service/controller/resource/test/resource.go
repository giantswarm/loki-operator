package test

import (
	"fmt"
	"strings"

	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/loki-operator/service/controller/promtailconfig"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Name = "todo"
	PromtailConfigLabel        = "giantswarm.io/loki-promtail-config"
	PromtailContainerNameLabel = "giantswarm.io/loki-promtail-container"
	PromtailConfigMapKeyName   = "promtail.yaml"
)

type Config struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
	Handler   promtailconfig.Handler
}

type Resource struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
	handler   promtailconfig.Handler
}

func New(config Config) (*Resource, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Handler == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Handler must not be empty", config)
	}

	r := &Resource{
		logger:    config.Logger,
		k8sClient: config.K8sClient,
		handler:   config.Handler,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) configKeyName(pod *v1.Pod) (*promtailconfig.Key, error) {
	containerName, found := pod.ObjectMeta.Labels[PromtailContainerNameLabel]
	if !found {
		if len(pod.Spec.Containers) != 1 {
			return nil, microerror.Maskf(invalidDynamicConfigError, "More than one container running in the Pod %s,"+
				" but no single logging container configure with '%s' Label", pod.Name, PromtailContainerNameLabel)
		}
		containerName = pod.Spec.Containers[0].Name
	} else {
		found := false
		for _, c := range pod.Spec.Containers {
			if c.Name == containerName {
				found = true
				break
			}
		}
		if !found {
			return nil, microerror.Maskf(invalidDynamicConfigError, "Container %v not found in pod %v, but "+
				"configured as the logging container of the pod using '%s' Label", containerName, pod.ObjectMeta.Name,
				PromtailContainerNameLabel)
		}
	}

	var labels strings.Builder
	for k, v := range pod.ObjectMeta.Labels {
		labels.WriteString(fmt.Sprintf("%v=%v,", k, v))
	}

	return &promtailconfig.Key{
		Namespace:     pod.Namespace,
		Labels:        labels.String(),
		ContainerName: containerName,
	}, nil
}

func (r *Resource) loadConfigMapByPod(pod *v1.Pod) (string, error) {
	namespace := pod.Namespace
	name, found := pod.ObjectMeta.Labels[PromtailConfigLabel]
	if !found {
		return "", microerror.Maskf(invalidDynamicConfigError, "Pod %s/%s doesn't have %s Label", pod.Namespace,
			pod.Name, PromtailConfigLabel)
	}

	cm, err := r.k8sClient.K8sClient().CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return "", microerror.Maskf(invalidDynamicConfigError, "Promtail ConfigMap named '%s' configured, but not found: %v",
			name, err)
	}
	cfgTxt, found := cm.Data[PromtailConfigMapKeyName]
	if !found {
		return "", microerror.Maskf(invalidDynamicConfigError, "'%s' key not found in ConfigMap named '%v' configured",
			name)
	}
	return cfgTxt, nil
}
