package backup

import (
	"os"

	"github.com/Netcracker/qubership-mongodb-supplementary/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-supplementary/pkg/utils"
	cUtils "github.com/Netcracker/qubership-nosqldb-operator-core/pkg/utils"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func BackupDeploymentTemplate(spec *v1alpha1.MongodbSupplServiceSpec, pvcName string, namespace string, nodeSelector map[string]string,
	env []v1.EnvVar, tolerations []v1.Toleration, storageDirectory string, emptyDir bool, numberOfReplicas int32, priorityClassName string, affinity *v1.Affinity, volumeMounts []v1.VolumeMount, volumes []v1.Volume) *v12.Deployment {
	secret := utils.MongoSecret
	storage := utils.BackupStorage
	uriScheme := utils.GetUriScheme(spec.TLS.Enabled)
	var volumeSource v1.VolumeSource
	if emptyDir {
		volumeSource = v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{
				Medium: "",
			},
		}
	} else {
		volumeSource = v1.VolumeSource{
			PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
				ClaimName: pvcName,
			},
		}
	}
	allowPrivilegeEscalation := false

	backupVolumeMounts := []v1.VolumeMount{
		{
			Name:      storage,
			MountPath: storageDirectory,
		},
		{
			Name:      secret,
			ReadOnly:  true,
			MountPath: "/opt/" + secret,
		},
	}

	backupVolumes := []v1.Volume{
		{
			Name:         storage,
			VolumeSource: volumeSource,
		},
		{
			Name: secret,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: secret,
				},
			},
		},
	}

	volumeMounts = append(backupVolumeMounts, volumeMounts...)
	volumes = append(backupVolumes, volumes...)

	dc := &v12.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      utils.BackupDaemon,
			Namespace: namespace,
			Labels: map[string]string{
				utils.App:                  utils.MongoCluster,
				utils.Microservice:         utils.BackupDaemon,
				utils.AppPartOf:            "mongodb-services",
				utils.Name:                 utils.BackupDaemon,
				utils.AppName:              utils.BackupDaemon,
				utils.AppInstance:          os.Getenv("RELEASE_NAME"),
				utils.AppVersion:           os.Getenv("APP_VERSION"),
				utils.AppComponent:         "backend",
				utils.AppManagedBy:         "operator",
				utils.AppManagedByOperator: "mongodb-services-operator",
				utils.AppTechnology:        "python",
			},
		},
		Spec: v12.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					utils.Name: utils.BackupDaemon,
				},
			},
			Strategy: v12.DeploymentStrategy{
				Type: v12.RecreateDeploymentStrategyType,
			},
			Replicas: &numberOfReplicas,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Labels: map[string]string{
						utils.Name: utils.BackupDaemon,
					},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: spec.ServiceAccountName,
					SecurityContext:    spec.PodSecurityContext,
					PriorityClassName:  priorityClassName,
					Containers: []v1.Container{
						v1.Container{
							Name:            utils.BackupDaemon,
							Image:           spec.Backup.DockerImage,
							ImagePullPolicy: spec.ImagePullPolicy,
							SecurityContext: &v1.SecurityContext{
								Capabilities: &v1.Capabilities{
									Drop: []v1.Capability{"ALL"},
								},
								AllowPrivilegeEscalation: &allowPrivilegeEscalation,
							},
							Ports: []v1.ContainerPort{
								v1.ContainerPort{
									Name:          "http",
									ContainerPort: cUtils.GetHTTPPort(spec.TLS.Enabled),
									Protocol:      "TCP",
								},
							},
							Env:       env,
							Resources: *spec.Backup.Resources,
							LivenessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									HTTPGet: &v1.HTTPGetAction{
										Path:   "/health",
										Port:   intstr.IntOrString{Type: intstr.Int, IntVal: cUtils.GetHTTPPort(spec.TLS.Enabled)},
										Scheme: uriScheme,
									},
								},
								InitialDelaySeconds: 5,
								TimeoutSeconds:      30,
								PeriodSeconds:       5,
								SuccessThreshold:    1,
								FailureThreshold:    12,
							},
							ReadinessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									HTTPGet: &v1.HTTPGetAction{
										Path:   "/health",
										Port:   intstr.IntOrString{Type: intstr.Int, IntVal: cUtils.GetHTTPPort(spec.TLS.Enabled)},
										Scheme: uriScheme,
									},
								},
								InitialDelaySeconds: 5,
								TimeoutSeconds:      30,
								PeriodSeconds:       5,
								SuccessThreshold:    1,
								FailureThreshold:    12,
							},
							VolumeMounts: volumeMounts,
						},
					},
					NodeSelector: nodeSelector,
					Affinity:     affinity,
					Tolerations:  tolerations,
					Volumes:      volumes,
				},
			},
		},
	}

	if spec.Backup.S3.SslVerify {

		dc.Spec.Template.Spec.Volumes = append(dc.Spec.Template.Spec.Volumes,
			v1.Volume{
				Name: "s3-ssl-certs",
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: spec.Backup.S3.SslSecretName,
					},
				},
			},
		)

		dc.Spec.Template.Spec.Containers[0].VolumeMounts = append(dc.Spec.Template.Spec.Containers[0].VolumeMounts,
			v1.VolumeMount{
				Name:      "s3-ssl-certs",
				ReadOnly:  true,
				MountPath: "/s3Certs",
			},
		)
	}
	cUtils.TLSSpecUpdate(&dc.Spec.Template.Spec, utils.RootCertPath, spec.TLS)
	cUtils.TLSServerSpecUpdate(&dc.Spec.Template.Spec, spec.TLS, spec.Backup.TLS.BackupDaemonCASecretName, utils.ServerCertsPath)

	return dc
}
