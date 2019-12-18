package test

import (
	"context"

	v1 "k8s.io/api/core/v1"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	pod, castOk := obj.(*v1.Pod)
	if !castOk {
		return nil
	}
	key, err := r.configKeyName(pod)
	if err != nil {
		return err
	}
	cfgTxt, err := r.loadConfigMapByPod(pod)
	if err != nil {
		return err
	}
	r.handler.AddConfig(*key, cfgTxt)
	return nil
}
