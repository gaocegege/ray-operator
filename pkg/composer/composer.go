package composer

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	rayv1alpha1 "github.com/gaocegege/ray-operator/api/v1alpha1"
	"github.com/gaocegege/ray-operator/pkg/consts"
)

const (
	ComposerName = "ray-operator-composer"
)

// Composer is the interface for composer.
type Composer interface {
	DesiredLauncher(ray *rayv1alpha1.Ray) (*appsv1.Deployment, error)
	DesiredConfigMap(ray *rayv1alpha1.Ray) (*corev1.ConfigMap, error)
}

type composer struct {
	record.EventRecorder
	scheme *runtime.Scheme
}

// New creates a new Composer.
func New(recorder record.EventRecorder, scheme *runtime.Scheme) Composer {
	return &composer{
		EventRecorder: recorder,
		scheme:        scheme,
	}
}

func (c composer) DesiredLauncher(ray *rayv1alpha1.Ray) (*appsv1.Deployment, error) {
	deploymentLabels := ray.Labels

	podLabels := getLauncherLabels(ray.Name, ray.Name)

	d := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ray.Name,
			Namespace: ray.Namespace,
			Labels:    deploymentLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: podLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:    podLabels,
					Namespace: ray.Namespace,
				},
				Spec: corev1.PodSpec{
					// TODO: Fix it, it is just a demo.
					ServiceAccountName: "autoscaler",
					Volumes: []corev1.Volume{
						{
							Name: consts.DefaultRayConfigMapVolume,
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: ray.Name,
									},
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            consts.DefaultRayLauncherName,
							Image:           consts.DefaultRayImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command: []string{
								// TODO: Fix it, it is just a demo. We need to
								// compose the yaml and use that one.
								"bash",
								"-c",
								fmt.Sprintf("echo yes | ray up %s", consts.DefaultRayConfigMapFilePath),
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      consts.DefaultRayConfigMapVolume,
									ReadOnly:  true,
									MountPath: consts.DefaultRayConfigMapMountPath,
								},
							},
						},
					},
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(ray, d, c.scheme); err != nil {
		return nil, err
	}
	return d, nil
}

func getLauncherLabels(rayName, rayLauncherName string) map[string]string {
	return map[string]string{
		consts.LabelRayLauncher: rayLauncherName,
		consts.LabelRay:         rayName,
	}
}
