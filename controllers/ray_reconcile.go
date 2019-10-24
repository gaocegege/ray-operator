package controllers

import (
	ctrl "sigs.k8s.io/controller-runtime"

	rayv1alpha1 "github.com/gaocegege/ray-operator/api/v1alpha1"
)

func (r *RayReconciler) reconcile(ray *rayv1alpha1.Ray) (ctrl.Result, error) {
	launcher, err := r.DesiredLauncher(ray)
	if err != nil {
		return ctrl.Result{}, err
	}

	_, err = r.createOrUpdateDeployment(ray, launcher)
	if err != nil {
		return ctrl.Result{}, err
	}

	// TODO: Update status.
	return ctrl.Result{}, nil
}
