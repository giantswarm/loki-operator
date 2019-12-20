package test

import (
	"context"

	v1 "k8s.io/api/core/v1"
)

func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	pod, castOk := obj.(*v1.Pod)
	if !castOk {
		return nil
	}
	key, err := r.configKeyName(pod)
	if err != nil {
		return nil
	}
	r.handler.DelConfig(*key)
	return nil
}
