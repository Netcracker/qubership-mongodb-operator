package robotTests

import (
	"github.com/Netcracker/qubership-mongodb-supplementary/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/types"
	utils2 "github.com/Netcracker/qubership-nosqldb-operator-core/pkg/utils"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ServiceName = "robot-tests"
	Name        = "name"
)

// TODO to core
func RobotTemplate(namespace string, image string,
	resources corev1.ResourceRequirements,
	nodeSelector map[string]string, tolerations []corev1.Toleration,
	env []corev1.EnvVar, securityContext *corev1.PodSecurityContext,
	tls types.TLS, vaultRegistration types.VaultRegistration, priorityClassName, serviceAccountName string, affinity *corev1.Affinity, imagePullPolicy corev1.PullPolicy, volumeMounts []corev1.VolumeMount, volumes []corev1.Volume) *v1.Deployment {
	allowPrivilegeEscalation := false
	readOnlyRootFilesystem := true
	var replicas int32 = 1

	volumes = append(volumes, corev1.Volume{
		Name: "tmp",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{
				SizeLimit: resource.NewScaledQuantity(32, resource.Mega),
			},
		},
	})

	volumeMounts = append(volumeMounts, corev1.VolumeMount{
		Name:      "tmp",
		MountPath: "/tmp",
	})

	dc := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ServiceName,
			Namespace: namespace,
			Labels: map[string]string{
				"app": ServiceName,
			},
		},
		Spec: v1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					Name: ServiceName,
				},
			},
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Labels: map[string]string{
						Name: ServiceName,
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: serviceAccountName,
					RestartPolicy:      corev1.RestartPolicyAlways,
					SecurityContext:    securityContext,
					PriorityClassName:  priorityClassName,
					Containers: []corev1.Container{
						corev1.Container{
							Name:            utils.Robot,
							Image:           image,
							Env:             env,
							VolumeMounts:    volumeMounts,
							Resources:       resources,
							ImagePullPolicy: imagePullPolicy,
							SecurityContext: &corev1.SecurityContext{
								ReadOnlyRootFilesystem: &readOnlyRootFilesystem,
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{"ALL"},
								},
								AllowPrivilegeEscalation: &allowPrivilegeEscalation,
							},
						},
					},
					NodeSelector: nodeSelector,
					Volumes:      volumes,
					Affinity:     affinity,
					Tolerations:  tolerations,
				},
			},
		},
	}

	utils2.VaultPodSpec(&dc.Spec.Template.Spec, []string{"/docker-entrypoint.sh", "run-robot"}, vaultRegistration)

	utils2.TLSSpecUpdate(&dc.Spec.Template.Spec, utils.RootCertPath, tls)

	return dc
}
