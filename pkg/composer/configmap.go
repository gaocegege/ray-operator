package composer

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	rayv1alpha1 "github.com/gaocegege/ray-operator/api/v1alpha1"
	"github.com/gaocegege/ray-operator/pkg/consts"
)

func (c composer) DesiredConfigMap(ray *rayv1alpha1.Ray) (*corev1.ConfigMap, error) {
	cmLabels := ray.Labels

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ray.Name,
			Namespace: ray.Namespace,
			Labels:    cmLabels,
		},
		Data: map[string]string{},
	}
	if ray.Spec.Configuration != nil {
		cm.Data[consts.DefaultRayConfigMapFile] = *ray.Spec.Configuration
	}

	// TODO: Deal with the inline configuration.

	if err := controllerutil.SetControllerReference(ray, cm, c.scheme); err != nil {
		return nil, err
	}
	return cm, nil
}
