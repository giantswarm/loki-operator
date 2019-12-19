package controller

import (
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/resource"
	"github.com/giantswarm/operatorkit/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/retryresource"
	v1 "k8s.io/api/core/v1"
	"time"

	"github.com/giantswarm/loki-operator/service/controller/promtailconfig"
	"github.com/giantswarm/loki-operator/service/controller/resource/test"
)

type todoResourceSetConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
	Handler   promtailconfig.Handler
}

func newTODOResourceSet(config todoResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	var testResource resource.Interface
	{
		pc, err := promtailconfig.NewPromtailConfigMap(config.K8sClient, "loki", "loki-promtail",
			test.PromtailConfigMapKeyName)
		if err != nil {
			return nil, err
		}
		handler, err := promtailconfig.NewPeriodicHandler(3*time.Second, 30*time.Second, pc)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		c := test.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
			Handler:   handler,
		}

		testResource, err = test.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		testResource,
	}

	{
		c := retryresource.WrapConfig{
			Logger: config.Logger,
		}

		resources, err = retryresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	{
		c := metricsresource.WrapConfig{}

		resources, err = metricsresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	handlesFunc := func(obj interface{}) bool {
		pod, castOk := obj.(*v1.Pod)
		if !castOk {
			return false
		}
		_, found := pod.ObjectMeta.Labels[test.PromtailConfigLabel]
		if !found {
			return false
		}
		return true
	}

	var resourceSet *controller.ResourceSet
	{
		c := controller.ResourceSetConfig{
			Handles:   handlesFunc,
			Logger:    config.Logger,
			Resources: resources,
		}

		resourceSet, err = controller.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resourceSet, nil
}
