package test

import (
	"context"

	v1 "k8s.io/api/core/v1"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	_, castOk := obj.(*v1.Pod)
	if !castOk {
		return nil
	}
	return nil
}
