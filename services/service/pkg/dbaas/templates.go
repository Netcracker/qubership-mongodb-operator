package dbaas

import (
	"github.com/Netcracker/qubership-mongodb-supplementary/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-supplementary/pkg/utils"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func DbaasDeploymentTemplate(spec *v1alpha1.MongodbSupplServiceSpec, namespace string, image string,
	nodeSelector map[string]string, resources v1.ResourceRequirements,
	env []v1.EnvVar, tolerations []v1.Toleration, numberOfReplicas int32, port int32, priorityClassName string, affinity *v1.Affinity) *v12.Deployment {

	allowPrivilegeEscalation := false
	dc := &v12.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      utils.DbaasName,
			Namespace: namespace,
			Labels: map[string]string{
				utils.App:          utils.MongoCluster,
				utils.Microservice: utils.DbaasName,
			},
		},
		Spec: v12.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					utils.Name: utils.DbaasName,
				},
			},
			Replicas: &numberOfReplicas,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Labels: map[string]string{
						utils.Name: utils.DbaasName,
					},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: spec.ServiceAccountName,
					SecurityContext:    spec.PodSecurityContext,
					PriorityClassName:  priorityClassName,
					Containers: []v1.Container{
						v1.Container{
							Name:            utils.DbaasName,
							Image:           image,
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
									ContainerPort: port,
									Protocol:      "TCP",
								},
							},
							Env:       env,
							Resources: resources,
							LivenessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									TCPSocket: &v1.TCPSocketAction{
										Port: intstr.IntOrString{Type: intstr.Int, IntVal: port},
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
									TCPSocket: &v1.TCPSocketAction{
										Port: intstr.IntOrString{Type: intstr.Int, IntVal: port},
									},
								},
								InitialDelaySeconds: 5,
								TimeoutSeconds:      30,
								PeriodSeconds:       5,
								SuccessThreshold:    1,
								FailureThreshold:    12,
							},
						},
					},
					NodeSelector: nodeSelector,
					Affinity:     affinity,
					Tolerations:  tolerations,
				},
			},
		},
	}

	return dc
}
