package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	rayv1alpha1 "github.com/gaocegege/ray-operator/api/v1alpha1"
)

func (r *RayReconciler) updateStatus(ray *rayv1alpha1.Ray) error {
	pods := &corev1.PodList{}
	if err := r.List(context.TODO(), pods,
		client.InNamespace(ray.Namespace)); err != nil {
		return err
	}
	return nil
}
