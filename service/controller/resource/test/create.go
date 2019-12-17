package test

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	pod, castOk := obj.(*v1.Pod)
	if !castOk {
		return nil
	}
	fmt.Print(pod)
	return nil
}
