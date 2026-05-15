package utils

import (
	"fmt"

	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	cUtils "github.com/Netcracker/qubership-nosqldb-operator-core/pkg/utils"
	v12 "k8s.io/api/apps/v1"
	corev12 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v13 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func SecretTemplate(name string, values map[string]string, namespace string) *v1.Secret {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: make(map[string]string),
		},
	}
	data := map[string][]byte{}
	for key, value := range values {
		data[key] = []byte(value)
	}

	secret.Data = data

	return secret
}

func ServiceAccountSecretTemplate(name string, saName string, namespace string) *v1.Secret {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: map[string]string{
				"kubernetes.io/service-account.name": saName,
			},
		},
		Type: v1.SecretTypeServiceAccountToken,
	}

	return secret
}

func MongoSSCommonTemplate(
	image string,
	resources v1.ResourceRequirements,
	affinity v1.Affinity,
	nodeSelector map[string]string,
	namespace string,
	pvcName string,
	mongoSecretName string,
	nameKey string,
	nameWithIndexes string,
	containerArgs string,
	securityContext *v1.PodSecurityContext,
	tolerations []v1.Toleration,
	mongoImagePullPolicy v1.PullPolicy,
	containerTimeoutSeconds int,
	containerPeriodSeconds int,
	tls v1alpha1.TLS,
	priorityClassName string) *v12.StatefulSet {

	statefulSet := MongoSSCommonTemplateWithoutNamesArgs(
		image,
		resources,
		affinity,
		nodeSelector,
		namespace,
		pvcName,
		mongoSecretName,
		mongoImagePullPolicy,
		containerTimeoutSeconds,
		containerPeriodSeconds,
		tls,
		priorityClassName)

	statefulSet.ObjectMeta.Name = nameWithIndexes
	statefulSet.Spec.ServiceName = nameKey

	statefulSet.ObjectMeta.Labels = map[string]string{Microservice: nameKey}

	statefulSet.Spec.Selector.MatchLabels[Microservice] = nameKey
	statefulSet.Spec.Selector.MatchLabels[MongoNode] = nameWithIndexes

	statefulSet.Spec.Template.ObjectMeta.Labels[Microservice] = nameKey
	statefulSet.Spec.Template.ObjectMeta.Labels[MongoNode] = nameWithIndexes

	statefulSet.Spec.Template.Spec.Containers[0].Name = nameWithIndexes

	statefulSet.Spec.Template.Spec.Containers[0].Args = []string{
		"-c",
		containerArgs,
	}

	statefulSet.Spec.Template.Spec.SecurityContext = securityContext
	statefulSet.Spec.Template.Spec.Tolerations = tolerations

	return statefulSet
}

func MongoSSCommonTemplateWithoutNamesArgs(
	image string,
	resources v1.ResourceRequirements,
	affinity v1.Affinity,
	nodeSelector map[string]string,
	namespace string,
	pvcName string,
	mongoSecretName string,
	mongoImagePullPolicy v1.PullPolicy,
	containerTimeoutSeconds int,
	containerPeriodSeconds int,
	tls v1alpha1.TLS,
	priorityClassName string) *v12.StatefulSet {

	mongoCluster := MongoCluster
	app := App
	data := Data
	replicas := int32(1)
	termination := int64(0)
	secretVolumeMode := int32(256)
	timeoutSeconds := int32(containerTimeoutSeconds)
	periodSeconds := int32(containerPeriodSeconds)
	allowPrivilegeEscalation := false
	readOnlyRootFilesystem := true

	statefulSet := v12.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Labels: map[string]string{
				CloneModeType: "data",
			},
		},
		Spec: v12.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					app:           mongoCluster,
					CloneModeType: "data",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						app:           mongoCluster,
						CloneModeType: "data",
					},
				},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						v1.Volume{
							Name: data,
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvcName,
								},
							},
						},
						v1.Volume{
							Name: mongoSecretName,
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName:  mongoSecretName,
									DefaultMode: &secretVolumeMode,
								},
							},
						},
						v1.Volume{
							Name: "tmp",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{
									SizeLimit: resource.NewScaledQuantity(32, resource.Mega),
								},
							},
						},
					},
					Containers: []v1.Container{
						v1.Container{
							Image:           image,
							ImagePullPolicy: mongoImagePullPolicy,
							SecurityContext: &v1.SecurityContext{
								ReadOnlyRootFilesystem: &readOnlyRootFilesystem,
								Capabilities: &v1.Capabilities{
									Drop: []v1.Capability{"ALL"},
								},
								AllowPrivilegeEscalation: &allowPrivilegeEscalation,
							},
							Command: []string{
								BashCommand,
							},
							ReadinessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									Exec: &v1.ExecAction{
										Command: []string{
											MongoBinary(image),
											"admin",
											"--host",
											"localhost",
											"--port",
											"27017",
											"--eval",
											"\"quit()\"",
										},
									},
								},
								InitialDelaySeconds: int32(10),
								TimeoutSeconds:      timeoutSeconds,
								PeriodSeconds:       periodSeconds,
							},
							Ports: []v1.ContainerPort{
								v1.ContainerPort{
									Name:          "mongo",
									ContainerPort: 27017,
									Protocol:      "TCP",
								},
							},
							VolumeMounts: []v1.VolumeMount{
								v1.VolumeMount{
									Name:      data,
									MountPath: "/" + data,
								},
								v1.VolumeMount{
									Name:      mongoSecretName,
									ReadOnly:  true,
									MountPath: "/opt/" + mongoSecretName,
								},
								v1.VolumeMount{
									Name:      "tmp",
									MountPath: "/tmp",
								},
							},
							Resources: resources,
						},
					},
					TerminationGracePeriodSeconds: &termination,
					Affinity:                      &affinity,
					NodeSelector:                  nodeSelector,
					PriorityClassName:             priorityClassName,
				},
			},
		},
	}

	TLSClientSpecUpdate(&statefulSet.Spec.Template.Spec, tls)

	return &statefulSet
}

func ShardServiceTemplate(name string, namespace string) *v1.Service {
	mongoCluster := MongoCluster
	app := App
	microservice := Microservice

	selectors := map[string]string{
		app:          mongoCluster,
		microservice: name,
	}

	labels := map[string]string{
		app:                                     mongoCluster,
		microservice:                            name,
		"app.kubernetes.io/part-of":             "mongodb",
		"name":                                  name,
		"app.kubernetes.io/name":                name,
		"app.kubernetes.io/managed-by":          "operator",
		"app.kubernetes.io/managed-by-operator": "mongodb-operator",
	}

	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name: "mongo",
					Port: 27017,
				},
			},
			Selector:  selectors,
			ClusterIP: "None",
		},
	}

}

func MongosServiceTemplate(name string, selector string, namespace string) *v1.Service {
	mongoCluster := MongoCluster
	app := App
	nameSelector := Name

	labels := map[string]string{
		app:                                     mongoCluster,
		"app.kubernetes.io/part-of":             "mongodb",
		"name":                                  name,
		"app.kubernetes.io/name":                name,
		"app.kubernetes.io/managed-by":          "operator",
		"app.kubernetes.io/managed-by-operator": "mongodb-operator",
	}

	selectors := map[string]string{
		nameSelector: selector,
	}

	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Name:       "mongo",
					Port:       27017,
					TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: 27017},
				},
			},
			Selector:        selectors,
			SessionAffinity: v1.ServiceAffinityClientIP,
		},
	}

}

func SingleMongosRCTemplate(
	image string,
	resources v1.ResourceRequirements,
	namespace string,
	containerArgs string,
	pvcName string,
	nodeSelector map[string]string,
	securityContext *v1.PodSecurityContext,
	tolerations []v1.Toleration,
	mongoImagePullPolicy v1.PullPolicy,
	tls v1alpha1.TLS,
	priorityClassName string,
	affinity *v1.Affinity) *v1.ReplicationController {

	numberOfReplicas := 1

	template := MongosRCTemplate(
		image,
		resources,
		namespace,
		containerArgs,
		securityContext,
		tolerations,
		mongoImagePullPolicy,
		numberOfReplicas,
		tls,
		priorityClassName,
		affinity,
	)

	template.Spec.Template.Spec.Containers[0].VolumeMounts = append(
		template.Spec.Template.Spec.Containers[0].VolumeMounts,
		v1.VolumeMount{
			Name:      Data,
			MountPath: "/" + Data,
		})

	template.Spec.Template.Spec.Volumes = append(
		template.Spec.Template.Spec.Volumes,
		v1.Volume{
			Name: Data,
			VolumeSource: v1.VolumeSource{
				PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
					ClaimName: pvcName,
				},
			},
		})

	template.Spec.Template.Spec.NodeSelector = nodeSelector

	return template
}

func MongosRCTemplate(
	image string,
	resources v1.ResourceRequirements,
	namespace string,
	containerArgs string,
	securityContext *v1.PodSecurityContext,
	tolerations []v1.Toleration,
	mongoImagePullPolicy v1.PullPolicy,
	numberOfReplicas int,
	tls v1alpha1.TLS,
	priorityClassName string,
	affinity *v1.Affinity) *v1.ReplicationController {

	selector := map[string]string{
		Name: Mongos,
		App:  MongoCluster,
	}

	replicas := int32(numberOfReplicas)
	defaultMode := int32(256)
	secret := MongoSecret
	mongos := Mongos
	allowPrivilegeEscalation := false
	readOnlyRootFilesystem := true

	defaultNodeAffinity := &v1.NodeAffinity{
		PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
			{
				Weight: 1,
				Preference: v1.NodeSelectorTerm{
					MatchExpressions: []v1.NodeSelectorRequirement{
						{
							Key:      "mongo",
							Operator: "In",
							Values: []string{
								"mongos",
							},
						},
					},
				},
			},
		},
	}

	// Default anti-affinity rule (used only if user didn't provide one)
	defaultAffinity := &v1.Affinity{
		PodAntiAffinity: &v1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
				v1.PodAffinityTerm{
					LabelSelector: &v13.LabelSelector{
						MatchExpressions: []v13.LabelSelectorRequirement{
							v13.LabelSelectorRequirement{
								Key:      Name,
								Operator: "In",
								Values: []string{
									mongos,
								},
							},
						},
					},
					TopologyKey: KubeHostName,
				},
			},
		},
		NodeAffinity: defaultNodeAffinity,
	}

	var finalAffinity *corev12.Affinity
	// Use user-defined affinity if available; otherwise use default
	if affinity != nil {
		finalAffinity = affinity
	} else {
		finalAffinity = defaultAffinity
	}

	template := &v1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mongos,
			Namespace: namespace,
			Labels:    selector,
		},
		Spec: v1.ReplicationControllerSpec{
			Replicas: &replicas,
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: selector,
				},
				Spec: v1.PodSpec{
					SecurityContext:   securityContext,
					PriorityClassName: priorityClassName,
					Tolerations:       tolerations,
					Affinity:          finalAffinity,
					Containers: []v1.Container{
						v1.Container{
							Name:            mongos,
							Image:           image,
							ImagePullPolicy: mongoImagePullPolicy,
							SecurityContext: &v1.SecurityContext{
								ReadOnlyRootFilesystem: &readOnlyRootFilesystem,
								Capabilities: &v1.Capabilities{
									Drop: []v1.Capability{"ALL"},
								},
								AllowPrivilegeEscalation: &allowPrivilegeEscalation,
							},
							Command: []string{
								BashCommand,
							},
							Args: []string{
								"-c",
								containerArgs,
							},
							ReadinessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									Exec: &v1.ExecAction{
										Command: []string{
											MongoBinary(image),
											"--port",
											"27017",
											"--eval",
											"quit()",
										},
									},
								},
								FailureThreshold:    int32(10),
								InitialDelaySeconds: int32(10),
								PeriodSeconds:       int32(10),
								SuccessThreshold:    int32(1),
								TimeoutSeconds:      int32(10),
							},
							Ports: []v1.ContainerPort{
								v1.ContainerPort{
									Name:          "mongo",
									ContainerPort: 27017,
									Protocol:      "TCP",
								},
							},
							VolumeMounts: []v1.VolumeMount{
								v1.VolumeMount{
									Name:      secret,
									ReadOnly:  true,
									MountPath: "/opt/" + secret,
								},
								v1.VolumeMount{
									Name:      "tmp",
									MountPath: "/tmp",
								},
							},
							Resources: resources,
						},
					},
					Volumes: []v1.Volume{
						v1.Volume{
							Name: secret,
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName:  secret,
									DefaultMode: &defaultMode,
								},
							},
						},
						v1.Volume{
							Name: "tmp",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{
									SizeLimit: resource.NewScaledQuantity(32, resource.Mega),
								},
							},
						},
					},
				},
			},
		},
	}

	TLSClientSpecUpdate(&template.Spec.Template.Spec, tls)

	return template
}

func RecyclerPodTemplate(pvcName string, namespace string, image string, nodeSelector map[string]string, serviceAccountName string, res v1.ResourceRequirements, securityContext *v1.PodSecurityContext) *v1.Pod {
	podName := fmt.Sprintf(RecyclerNameTemplate, pvcName)
	allowPrivilegeEscalation := false
	readOnlyRootFilesystem := true
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
			Labels: map[string]string{
				App:          MongoCluster,
				Microservice: RecyclerPod,
			},
		},
		Spec: v1.PodSpec{
			ServiceAccountName: serviceAccountName,
			SecurityContext:    securityContext,
			RestartPolicy:      v1.RestartPolicyNever,
			Volumes: []v1.Volume{
				v1.Volume{
					Name: podName,
					VolumeSource: v1.VolumeSource{
						PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
							ClaimName: pvcName,
						},
					},
				},
			},
			Containers: []v1.Container{
				v1.Container{
					Name:  fmt.Sprintf(RecyclerNameTemplate, "container"),
					Image: image,
					SecurityContext: &v1.SecurityContext{
						ReadOnlyRootFilesystem: &readOnlyRootFilesystem,
						Capabilities: &v1.Capabilities{
							Drop: []v1.Capability{"ALL"},
						},
						AllowPrivilegeEscalation: &allowPrivilegeEscalation,
					},
					Command: []string{
						"/bin/sh",
						"-c",
						"echo \"clearing pvc\" && ls -lah /scrub && rm -rf /scrub/*  && test -z \"$(ls -A /scrub)\" && ls -lah /scrub || exit 1",
					},
					VolumeMounts: []v1.VolumeMount{
						v1.VolumeMount{
							Name:      podName,
							MountPath: "/scrub",
						},
					},
					Resources: res,
				},
			},
			NodeSelector: nodeSelector,
		},
	}

	return pod
}

func SimpleServiceTemplate(name string, port int32, namespace string) *v1.Service {
	servicePort := []v1.ServicePort{
		v1.ServicePort{
			Name:       "http",
			Port:       port,
			TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: port},
		},
	}
	mongoCluster := MongoCluster
	app := App
	nameSelector := Name

	labels := map[string]string{
		app:          mongoCluster,
		Microservice: name,
	}

	selectors := map[string]string{
		nameSelector: name,
	}

	return cUtils.MultiportServiceTemplate(name, labels, selectors, &servicePort, namespace)
}
