// Copyright 2024-2025 NetCracker Technology Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package prometheus_exporter

import (
	"os"

	"github.com/Netcracker/qubership-mongodb-supplementary/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/types"
	cUtils "github.com/Netcracker/qubership-nosqldb-operator-core/pkg/utils"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func PrometheusExporterDeploymentTemplate(namespace string, image string, nodeSelector map[string]string, serviceAccountName string,
	resources v1.ResourceRequirements, env []v1.EnvVar, securityContext *v1.PodSecurityContext, tolerations []v1.Toleration, prometheusExporterImagePullPolicy v1.PullPolicy,
	tls types.TLS, priorityClassName, instance string, healthzPort int, volumeMounts []v1.VolumeMount, volumes []v1.Volume) *v12.Deployment {
	var replicas int32 = 1
	allowPrivilegeEscalation := false
	readOnlyRootFilesystem := true
	dc := &v12.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      utils.MongoPrometheusExporter,
			Namespace: namespace,
			Labels: map[string]string{
				utils.Name:                 utils.MongoPrometheusExporter,
				utils.AppPartOf:            "mongodb-services",
				utils.AppName:              utils.MongoPrometheusExporter,
				utils.AppInstance:          os.Getenv("RELEASE_NAME"),
				utils.AppVersion:           os.Getenv("APP_VERSION"),
				utils.AppComponent:         "backend",
				utils.AppManagedBy:         "operator",
				utils.AppManagedByOperator: "mongodb-services-operator",
				utils.AppTechnology:        "go",
			},
		},
		Spec: v12.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					utils.Name: utils.MongoPrometheusExporter,
				},
			},
			Replicas: &replicas,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Labels: map[string]string{
						utils.Name: utils.MongoPrometheusExporter,
					},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: serviceAccountName,
					SecurityContext:    securityContext,
					PriorityClassName:  priorityClassName,
					Tolerations:        tolerations,
					Containers: []v1.Container{
						v1.Container{
							Name:            utils.MongoPrometheusExporter,
							Image:           image,
							ImagePullPolicy: prometheusExporterImagePullPolicy,
							Command:         []string{"bash", "-c", "/opt/run.sh"},
							Env:             env,
							VolumeMounts:    volumeMounts,
							Resources:       resources,
							SecurityContext: &v1.SecurityContext{
								ReadOnlyRootFilesystem: &readOnlyRootFilesystem,
								Capabilities: &v1.Capabilities{
									Drop: []v1.Capability{"ALL"},
								},
								AllowPrivilegeEscalation: &allowPrivilegeEscalation,
							},
							LivenessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									TCPSocket: &v1.TCPSocketAction{
										Port: intstr.IntOrString{Type: intstr.Int, IntVal: int32(healthzPort)},
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
										Port: intstr.IntOrString{Type: intstr.Int, IntVal: int32(healthzPort)},
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
					Volumes:      volumes,
					NodeSelector: nodeSelector,
				},
			},
		},
	}

	cUtils.TLSSpecUpdate(&dc.Spec.Template.Spec, utils.RootCertPath, tls)

	return dc
}
