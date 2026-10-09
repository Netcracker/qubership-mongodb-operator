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

package main

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"go.uber.org/zap"
	gock "gopkg.in/h2non/gock.v1"
	v12 "k8s.io/api/apps/v1"

	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-operator/pkg"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	cTypes "github.com/Netcracker/qubership-nosqldb-operator-core/pkg/types"
	cUtils "github.com/Netcracker/qubership-nosqldb-operator-core/pkg/utils"
	v1core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type TestUtilsImpl struct {
	core.DefaultKubernetesHelperImpl
}

func (r *TestUtilsImpl) OpensslCommand(arg []string) ([]byte, error) {
	return []byte(
		"aEw5Zd++lhyFy6Z5z5NcrsZcheVc6Lv/vvrTVLmgBOD1qTZ8JKuam6e+fDcxTUeW" +
			"vKu9e+j40Tn5gJYTr6WvyZNW8bPIq5RaB/wHYsliY2xikkSC0R9PORCNKdHL99AL" +
			"gDap8JU/kfLlcxnbsPK7YjDpTbcOC/IZBOg31ehhBg7FZ047OMzaIaN+RG7Uik/D" +
			"fzHswBlzhfaliVIbMhGF+5vvInJKDEggCXb48s1+D9ke3WbF7vshkSJ8J1Xi3PH8" +
			"K2E36T8BQIl1megkm7b3fZx+jfAmQj29BKZIJ+DcPJUxlTK5mPQYcuyD2B+rqLYu" +
			"92Yq1glQNL0FZj4E/6FI9SyplPsIv4ipDZrJG3AYEGZwmYqOaB7k75RTaGPE33Nf" +
			"5+V6dNTcnx2e73EEPDv62raUPxcpw9BEE9oFotfHIcjDJSNPrj9S/N365jjy19fP" +
			"36FwFLQEI1hUGKI2EvrVYUUz8pmgyJH0iyfgIFqR1fUxHov/ncWKecUDkTn+D++a" +
			"8Aos3erifzy06D4/GA0jJ3ptt1j+3AUA0hhn/eU1pR6kQKcbR1jMxLcbWySM1cyD" +
			"MJW0Aor9fV1a11AIuCZZOEciID6mf3fxFeCXQJRgWmIYo4HJKG7GMQiS5MlD1Z6O" +
			"IFei//38aL7md4973agTI/QP+sUbV9ZhAxBFGt6dVolTGZ8XSIdmHldgFA4yO4Vh" +
			"bruA73uxJq2mQSuQW/EcUeCDKmJZwyWDOT0PeLFLXM5doMppTD4GXrQ1y/Hr3mT/" +
			"J423N+ipPiPK2FxpMu2vG9gj36L4blqR4PdDT1HKFBIoHs1/1uNarPSOvfLncHkD" +
			"nuEvE/1G51AOgcmiJG23faEdAI+8363M0w0SrUQbYTljLq5u0GsnXYsRcfxLWAyx" +
			"k5QjwRNpqUzfeijfnTRLcMMq+C9/+PwJt7WUHzKAXgMEoEM2ir/jnmIgLvWjip/5" +
			"ZZ3nH4ZtSS1Q2iN0jvSs/PxSxxmgtgXZgG5e7hCRuXvl1Jc="), nil
}

func (r *TestUtilsImpl) WaitForPVCBound(pvcName string, namespace string, waitSeconds int) error {
	return nil
}

func (r *TestUtilsImpl) WaitForDeploymentReady(deployName string, namespace string, waitSeconds int) error {
	return nil
}

func (r *TestUtilsImpl) WaitForPodsReady(labelSelectors map[string]string, namespace string, numberOfPods int, waitSeconds int) error {
	return nil
}

func (r *TestUtilsImpl) WaitForPodsCompleted(labelSelectors map[string]string, namespace string, numberOfPods int, waitSeconds int) error {
	return nil
}

func (r *TestUtilsImpl) WaitPodsCountByLabel(labelSelectors map[string]string, namespace string, numberOfPods int, waitSeconds int) error {
	return nil
}

func (r *TestUtilsImpl) ExecRemote(log *zap.Logger, kubeConfig *rest.Config, podName string, namespace string, containerName string, command string, args []string) (string, error) {
	for _, arg := range args {
		if strings.Contains(arg, "shouldBeHidden") {
			return "Ok", nil
		}
		if strings.Contains(arg, "shards.update") {
			return "1", nil
		}
		if arg == utils.RobotTestsFilesArgCheck {
			return utils.RobotTestsExpectedAnswer, nil
		}
	}

	return "", nil
}

func (r *TestUtilsImpl) WaitForTestsReady(deployName string, namespace string, waitSeconds int) error {
	return nil
}

func (r *TestUtilsImpl) UpdateRootPassword(authDB, username, oldPassword, newPassword string, inMonogs, inShards bool, shardsCound int) error {
	return nil
}

type TestMongoUtilsHelper struct {
	utils.MongoUtilsHelperImpl
	mongosPod       func(namespace string) *v1core.Pod
	primaryTemplate func(name string, namespace string) *v1core.Pod
}

var _ utils.MongoHelper = &TestMongoUtilsHelper{}

func (r *TestMongoUtilsHelper) SetMongoCMD(cmd string) {

}

func (r *TestMongoUtilsHelper) AddReplica(labels map[string]string, host string) error {
	return nil
}

func (r *TestMongoUtilsHelper) AddDRReplicas(labels map[string]string, rs []string) error {
	return nil
}

func (r *TestMongoUtilsHelper) GetMongoVersion() (string, error) {
	return "3.4.19", nil
}

func (r *TestMongoUtilsHelper) MongoReplicaSetInit(nameKey string, nameKeyWithIndex string, replicaSetSize int, namespace string,
	domain string, configInit bool, memberOptionsFunc func(i int) string) error {
	return nil
}

func (r *TestMongoUtilsHelper) GetMongosPod() (*v1core.Pod, error) {
	return r.mongosPod(r.Namespace), nil
}

func (r *TestMongoUtilsHelper) RunOnMongos(arg string) (string, error) {
	return r.RunOnMongoPod(
		r.mongosPod(r.Namespace),
		arg)
}

func (r *TestMongoUtilsHelper) GetMongoPrimaryReplica(labelSelectors map[string]string) (*v1core.Pod, error) {
	name := labelSelectors[utils.Microservice]
	return r.primaryTemplate(name, r.Namespace), nil
}

func (r *TestMongoUtilsHelper) RunOnPrimary(labels map[string]string, arg string) (string, error) {
	name := labels[utils.Microservice]
	return r.RunOnMongoPod(
		r.primaryTemplate(name, r.Namespace),
		arg)
}

func (r *TestMongoUtilsHelper) ReconfigureRS(labels map[string]string, replicasetHostnames []string) error {
	return nil
}

func (r *TestMongoUtilsHelper) RunOnShards(command string, shardCount int) ([]string, error) {
	return []string{}, nil
}

func (r *TestMongoUtilsHelper) CheckUserExists(user string) bool {
	return false
}

func (r *TestMongoUtilsHelper) CheckRoleExists(user string) bool {
	return false
}

func (r *TestMongoUtilsHelper) CreateRole(authDB, role, privileges, roles string, inMonogs, inShards bool, shardsCound int) error {
	return nil
}

func (r *TestMongoUtilsHelper) CreateUser(authDB, user, pass, role string, force, inMonogs, inShards bool, shardsCound int) error {
	return nil
}

func (r *TestMongoUtilsHelper) SetFeatureCompatibilityVersion(sharded bool, shardCount int) error {
	return nil
}

func (r *TestMongoUtilsHelper) GetClusterStatus(mode string, domain string, cnfReplicaSize int, dataReplicaSize int, shardCount int, sharded bool) (string, error) {
	return utils.Up, nil
}

func (r *TestMongoUtilsHelper) CheckReplicationLag(labels map[string]string, memberHostnames []string, maxLagSeconds int) (bool, error) {
	return true, nil
}

func getMaxReplica(schemaSettings v1alpha1.SchemaSettings) int {
	if schemaSettings.SchemaType == v1alpha1.Single {
		return 1
	} else {
		return core.MaxInt(schemaSettings.DataReplicaSize, schemaSettings.CnfReplicaSize)
	}
}

func genreateSecrets(namespace string) []runtime.Object {
	return []runtime.Object{
		cUtils.SecretTemplate(
			utils.BackupSecretName,
			map[string]string{
				utils.Username: "backup",
				utils.Password: "backup",
			},
			namespace),
		cUtils.SecretTemplate(
			utils.RestoreUserSecretName,
			map[string]string{
				utils.Username: "restore",
				utils.Password: "restore",
			},
			namespace),
		cUtils.SecretTemplate(
			utils.DbaasAdminSecretName,
			map[string]string{
				utils.Username: "dbaas",
				utils.Password: "dbaas",
			},
			namespace),
		cUtils.SecretTemplate(
			utils.DbaasAggregatorSecretName,
			map[string]string{
				utils.Username: "dbaas",
				utils.Password: "dbaas",
			},
			namespace),
		cUtils.SecretTemplate(
			utils.MonitoringSecretName,
			map[string]string{
				utils.Username: "monitoring",
				utils.Password: "monitoring",
			},
			namespace),
		cUtils.SecretTemplate(
			utils.MongoRootSecretName,
			map[string]string{
				utils.Username: "root",
				utils.Password: "root",
			},
			namespace),
		cUtils.SecretTemplate(
			utils.DbaasAdminRoleSecretName,
			map[string]string{
				"roles":      "{ role: 'read', db: 'admin' }, { role: 'read', db: 'config' }",
				"privileges": "{ resource: { cluster : true }, actions: ['listDatabases', 'find', 'changeStream'] }",
			},
			namespace),
	}
}

func generatePV(nameFormat string, nodeFormat string, namespace string, size int) ([]runtime.Object, []string, []map[string]string, []string) {
	pvS := []runtime.Object{}
	names := []string{}
	nodeLabels := []map[string]string{}
	pvSize := []string{}
	randomLabelKey := utils.RandomString(8)
	for i := 1; i <= size; i++ {
		pvS = append(pvS, &v1core.PersistentVolume{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf(nameFormat, i),
				Namespace: namespace,
				Labels: map[string]string{
					"node": fmt.Sprintf(nodeFormat, i),
				},
			},
		})
		names = append(names, fmt.Sprintf(nameFormat, i))
		nodeLabels = append(nodeLabels, map[string]string{randomLabelKey: fmt.Sprintf(nodeFormat, i)})
		pvSize = append(pvSize, "1Gi")
	}

	return pvS, names, nodeLabels, pvSize
}

func GenerateDefaultMongoService(namespace string, mongoPvNames []string, mongoNodes []map[string]string, mongoPvSize []string, backupPvName []string, backupNode []map[string]string, backupPvSize []string, schemaSettings v1alpha1.SchemaSettings) *v1alpha1.MongodbDeployment {
	GiQuantity, _ := resource.ParseQuantity("5Gi")
	var fsGroup int64 = 1001
	var tolerationSeconds int64 = 20

	rr := &v1core.ResourceRequirements{
		Limits: v1core.ResourceList{
			v1core.ResourceMemory: GiQuantity,
		},
		Requests: nil,
	}

	return &v1alpha1.MongodbDeployment{
		Spec: v1alpha1.MongodbDeploymentSpec{
			WaitSeconds:    100,
			AuthDb:         "admin",
			SchemaSettings: schemaSettings,
			Recycler: v1alpha1.Recycler{
				Resources: rr,
			},
			PodSecurityContext: &v1core.PodSecurityContext{
				FSGroup: &fsGroup,
			},
			Policies: &v1alpha1.Policies{
				Tolerations: []v1core.Toleration{
					{
						Key:               "key1",
						Value:             "value1",
						Operator:          v1core.TolerationOpEqual,
						Effect:            v1core.TaintEffectNoSchedule,
						TolerationSeconds: &tolerationSeconds,
					},
					{
						Key:               "key2",
						Value:             "value2",
						Operator:          v1core.TolerationOpEqual,
						Effect:            v1core.TaintEffectNoExecute,
						TolerationSeconds: &tolerationSeconds,
					},
				},
			},
			DisasterRecovery: &v1alpha1.DisasterRecovery{
				Mode:   "ACTIVE",
				Status: "norm",
				NoWait: false,
			},
			MongoDB: v1alpha1.MongoDB{
				Install: true,
				//RootUser:        "root",
				//RootPassword:    "root",
				CnfResources:    rr,
				DataResources:   rr,
				MongosResources: rr,
				Storage: &cTypes.StorageRequirements{
					Size:       mongoPvSize,
					Volumes:    mongoPvNames,
					NodeLabels: mongoNodes,
				},
				AdditionalNodeLabels: map[string]string{
					"common-label": "common-values",
				},
				MongoRootSecretName: "mongodb-root-credentials.v1",
			},
		},
	}
}

type CaseStruct struct {
	name                          string
	executor                      core.Executor
	builder                       core.ExecutableBuilder
	drBuilder                     core.ExecutableBuilder
	ctx                           core.ExecutionContext
	ctxToReplaceAfterServiceBuilt map[string]interface{}
	ReadResultFunc                func(t *testing.T, err error)
}

func GenerateDefaultMongoTestCase(testName string, mongoServiceSpec *v1alpha1.MongodbDeployment, runtimeObjects []runtime.Object,
	nameSpace string, nameSpaceRequestName string) CaseStruct {
	fakeClient := fake.NewFakeClient(runtimeObjects...)

	utilsHelp := &TestUtilsImpl{}
	utilsHelp.Client = fakeClient

	mongoHelp := &TestMongoUtilsHelper{
		mongosPod: func(namespace string) *v1core.Pod {
			return &v1core.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      utils.Mongos,
					Namespace: namespace,
				},
				Spec: v1core.PodSpec{
					Containers: []v1core.Container{
						{
							Name: utils.Mongos,
						},
					},
				},
			}
		},
		primaryTemplate: func(name string, namespace string) *v1core.Pod {
			return &v1core.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: v1core.PodSpec{
					Containers: []v1core.Container{
						{
							Name: name,
						},
					},
				},
			}
		},
	}
	mongoHelp.KubernetesHelperImpl = utilsHelp
	mongoHelp.Namespace = nameSpace
	mongoHelp.Client = fakeClient
	mongoHelp.Sharded = true
	mongoHelp.Single = true

	caseStruct := CaseStruct{
		name:      testName,
		executor:  core.DefaultExecutor(),
		builder:   &impl.MongoServiceBuilder{},
		drBuilder: &impl.DRBuilder{},
		ctx: core.GetExecutionContext(map[string]interface{}{
			constants.ContextSpec:   mongoServiceSpec,
			constants.ContextSchema: &runtime.Scheme{},
			constants.ContextRequest: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: nameSpace,
					Name:      nameSpaceRequestName,
				},
			},
			constants.ContextClient:        fakeClient,
			constants.ContextKubeClient:    &rest.Config{},
			constants.ContextLogger:        core.GetLogger(true),
			constants.ContextHashConfigMap: "random",
		}),
		ctxToReplaceAfterServiceBuilt: map[string]interface{}{
			utils.KubernetesHelperImpl: utilsHelp,
			utils.MongoHelperImpl:      mongoHelp,
		},
		ReadResultFunc: func(t *testing.T, err error) {
			if err != nil {
				// Some error happened
				t.Error(err)
			}
		},
	}

	caseStruct.ctx.Set(utils.MaxPVCCountForService, 3)

	return caseStruct
}

func GenerateDefaultMongoTestCaseWrapper(testName string, schemaSettings v1alpha1.SchemaSettings) CaseStruct {

	nameSpace := "mongo-namespace"
	nameSpaceRequestName := "mongo-name"

	mongoPvs, mongoPvNames, mongoNodes, mongoPvSize := generatePV("mongo-%v", "node-%v", nameSpaceRequestName, getMaxReplica(schemaSettings))
	backupPv, backupName, backupNodes, backupPvSize := generatePV("mongo-backup-%v", "node-%v", nameSpaceRequestName, 1)
	secrets := genreateSecrets(nameSpace)

	allPvs := append(mongoPvs, backupPv...)
	allObjects := append(allPvs, secrets...)

	mongoService := GenerateDefaultMongoService(
		nameSpace,
		mongoPvNames,
		mongoNodes,
		mongoPvSize,
		backupName,
		backupNodes,
		backupPvSize,
		schemaSettings,
	)

	return GenerateDefaultMongoTestCase(
		testName,
		mongoService,
		allObjects,
		nameSpace,
		nameSpaceRequestName,
	)
}

func TestExecutionCheck(t *testing.T) {
	gock.New("http://mongodb-operator.mongo-namespace.svc.cluster.local:8069").
		Get("/keyfile").
		Persist().
		Reply(200).
		JSON("key")

	testFuncs := []func() CaseStruct{
		func() CaseStruct {
			cs := GenerateDefaultMongoTestCaseWrapper(
				"Only Mongo HA Scheme",
				v1alpha1.SchemaSettings{
					SchemaType:      v1alpha1.HA,
					CnfReplicaSize:  3,
					DataReplicaSize: 3,
					ShardCount:      3,
					Sharded:         true,
				},
			)
			cs.executor.SetExecutable(cs.builder.Build(cs.ctx))
			msS := cs.ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
			cs.ctx.Set(constants.ContextSpec, msS)
			return cs
		},
		func() CaseStruct {
			cs := GenerateDefaultMongoTestCaseWrapper(
				"Only Mongo Single Scheme",
				v1alpha1.SchemaSettings{
					SchemaType: v1alpha1.Single,
				},
			)
			cs.executor.SetExecutable(cs.builder.Build(cs.ctx))
			msS := cs.ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
			cs.ctx.Set(constants.ContextSpec, msS)
			return cs
		},
		func() CaseStruct {
			cs := GenerateDefaultMongoTestCaseWrapper(
				"Mongo ipV6 HA Scheme",
				v1alpha1.SchemaSettings{
					SchemaType:      v1alpha1.HA,
					CnfReplicaSize:  3,
					DataReplicaSize: 3,
					ShardCount:      3,
					Sharded:         true,
				},
			)
			cs.executor.SetExecutable(cs.builder.Build(cs.ctx))
			msS := cs.ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
			msS.Spec.IpV6 = true
			cs.ctx.Set(constants.ContextSpec, msS)
			return cs
		},
		func() CaseStruct {
			cs := GenerateDefaultMongoTestCaseWrapper(
				"Only Mongo ipV6 Single Scheme",
				v1alpha1.SchemaSettings{
					SchemaType: v1alpha1.Single,
				},
			)
			cs.executor.SetExecutable(cs.builder.Build(cs.ctx))
			msS := cs.ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
			msS.Spec.IpV6 = true
			cs.ctx.Set(constants.ContextSpec, msS)
			return cs
		},
		func() CaseStruct {
			cnfrsCount := 3
			datarsCount := 3
			shardsCount := 3
			cs := GenerateDefaultMongoTestCaseWrapper(
				"All Services HA Scheme",
				v1alpha1.SchemaSettings{
					SchemaType:      v1alpha1.HA,
					CnfReplicaSize:  cnfrsCount,
					DataReplicaSize: datarsCount,
					ShardCount:      shardsCount,
					Sharded:         true,
				},
			)
			cs.executor.SetExecutable(cs.builder.Build(cs.ctx))
			msS := cs.ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
			mongoVols := msS.Spec.MongoDB.Storage.Volumes

			cs.ReadResultFunc = func(t *testing.T, err error) {
				cl := cs.ctx.Get(constants.ContextClient).(client.Client)
				request := cs.ctx.Get(constants.ContextRequest).(reconcile.Request)

				vcL := &v1core.PersistentVolumeClaimList{}
				listOps := []client.ListOption{
					client.InNamespace(request.Namespace),
				}
				cl.List(context.TODO(), vcL, listOps...)

				volClaimCheck := func(logName string, volume v1core.Volume, expectedClaimVolumeName string) {
					for _, vc := range vcL.Items {
						if vc.ObjectMeta.Name == volume.PersistentVolumeClaim.ClaimName {
							if vc.Spec.VolumeName != expectedClaimVolumeName {
								t.Error(fmt.Sprintf("%s's volume claim has worng volume name. Current is %s, expected %s",
									logName, vc.Spec.VolumeName, expectedClaimVolumeName))
								return
							} else {
								// Proper volume name
								t.Log("Found correct volume " + expectedClaimVolumeName + " for " + logName)
								return
							}
						}
					}
					t.Error(fmt.Sprintf("Not found PersistentVolumeClaim with %s name for %s", volume.PersistentVolumeClaim.ClaimName, logName))
				}

				backup := &v12.Deployment{}
				cl.Get(context.TODO(), types.NamespacedName{Name: utils.BackupDaemon, Namespace: request.Namespace}, backup)

				ssCheck := func(ssName string, expectedClaimVolumeName string) {
					ss := &v12.StatefulSet{}
					cl.Get(context.TODO(), types.NamespacedName{Name: ssName, Namespace: request.Namespace}, ss)
					volClaimCheck(ssName, ss.Spec.Template.Spec.Volumes[0], expectedClaimVolumeName)
				}

				for i := 0; i < cnfrsCount; i++ {
					nameWithIndex := fmt.Sprintf(utils.CnfNameWithIndexFormat, i)
					ssCheck(nameWithIndex, mongoVols[i])
				}
				for s := 0; s < shardsCount; s++ {
					for i := 0; i < datarsCount; i++ {
						nameWithIndexes := fmt.Sprintf(utils.DataNameWithIndexesFormat, s+1, i)
						ssCheck(nameWithIndexes, mongoVols[i])
					}
				}
			}
			return cs
		},
		func() CaseStruct {
			cs := GenerateDefaultMongoTestCaseWrapper(
				"All Services Single Scheme",
				v1alpha1.SchemaSettings{
					SchemaType: v1alpha1.Single,
				},
			)
			cs.executor.SetExecutable(cs.builder.Build(cs.ctx))
			return cs
		},
		func() CaseStruct {
			cs := GenerateDefaultMongoTestCaseWrapper(
				"All Services Arbiter Scheme",
				v1alpha1.SchemaSettings{
					SchemaType:      v1alpha1.Arbiter,
					CnfReplicaSize:  3,
					DataReplicaSize: 5,
					ShardCount:      3,
					Sharded:         true,
				},
			)
			cs.executor.SetExecutable(cs.builder.Build(cs.ctx))
			return cs
		},
		func() CaseStruct {
			cs := GenerateDefaultMongoTestCaseWrapper(
				"Only Backup Single Scheme",
				v1alpha1.SchemaSettings{
					SchemaType: v1alpha1.Single,
				},
			)
			cs.executor.SetExecutable(cs.builder.Build(cs.ctx))
			msS := cs.ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
			msS.Spec.MongoDB.Install = false
			cs.ctx.Set(constants.ContextSpec, msS)
			return cs
		},
		func() CaseStruct {
			cs := GenerateDefaultMongoTestCaseWrapper(
				"Only Backup HA Scheme",
				v1alpha1.SchemaSettings{
					SchemaType:      v1alpha1.HA,
					CnfReplicaSize:  3,
					DataReplicaSize: 3,
					ShardCount:      3,
					Sharded:         true,
				},
			)
			cs.executor.SetExecutable(cs.builder.Build(cs.ctx))
			msS := cs.ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
			msS.Spec.MongoDB.Install = false
			cs.ctx.Set(constants.ContextSpec, msS)
			return cs
		},
		func() CaseStruct {
			cs := GenerateDefaultMongoTestCaseWrapper(
				"Only Backup Arbiter Scheme",
				v1alpha1.SchemaSettings{
					SchemaType:      v1alpha1.Arbiter,
					CnfReplicaSize:  3,
					DataReplicaSize: 5,
					ShardCount:      3,
					Sharded:         true,
				},
			)
			cs.executor.SetExecutable(cs.builder.Build(cs.ctx))
			msS := cs.ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
			msS.Spec.MongoDB.Install = false
			cs.ctx.Set(constants.ContextSpec, msS)
			return cs
		},
		func() CaseStruct {
			cs := GenerateDefaultMongoTestCaseWrapper(
				"Only Dbaas HA Scheme",
				v1alpha1.SchemaSettings{
					SchemaType:      v1alpha1.HA,
					CnfReplicaSize:  3,
					DataReplicaSize: 3,
					ShardCount:      3,
					Sharded:         true,
				},
			)
			cs.executor.SetExecutable(cs.builder.Build(cs.ctx))
			msS := cs.ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
			msS.Spec.MongoDB.Install = false
			cs.ctx.Set(constants.ContextSpec, msS)
			return cs
		},
		func() CaseStruct {
			cs := GenerateDefaultMongoTestCaseWrapper(
				"Only Dbaas Arbiter Scheme",
				v1alpha1.SchemaSettings{
					SchemaType:      v1alpha1.Arbiter,
					CnfReplicaSize:  3,
					DataReplicaSize: 5,
					ShardCount:      3,
					Sharded:         true,
				},
			)
			cs.executor.SetExecutable(cs.builder.Build(cs.ctx))
			msS := cs.ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
			msS.Spec.MongoDB.Install = false
			cs.ctx.Set(constants.ContextSpec, msS)
			return cs
		},
		func() CaseStruct {
			cs := GenerateDefaultMongoTestCaseWrapper(
				"Only Dbaas Single Scheme",
				v1alpha1.SchemaSettings{
					SchemaType: v1alpha1.Single,
				},
			)
			cs.executor.SetExecutable(cs.builder.Build(cs.ctx))
			msS := cs.ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
			msS.Spec.MongoDB.Install = false
			cs.ctx.Set(constants.ContextSpec, msS)
			return cs
		},
		func() CaseStruct {
			cs := GenerateDefaultMongoTestCaseWrapper(
				"Only Monitoring HA Scheme",
				v1alpha1.SchemaSettings{
					SchemaType:      v1alpha1.HA,
					CnfReplicaSize:  3,
					DataReplicaSize: 3,
					ShardCount:      3,
					Sharded:         true,
				},
			)
			cs.executor.SetExecutable(cs.builder.Build(cs.ctx))
			msS := cs.ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
			msS.Spec.MongoDB.Install = false
			cs.ctx.Set(constants.ContextSpec, msS)
			return cs
		},
		func() CaseStruct {
			cs := GenerateDefaultMongoTestCaseWrapper(
				"Only Monitoring Arbiter Scheme",
				v1alpha1.SchemaSettings{
					SchemaType:      v1alpha1.Arbiter,
					CnfReplicaSize:  3,
					DataReplicaSize: 5,
					ShardCount:      3,
					Sharded:         true,
				},
			)
			cs.executor.SetExecutable(cs.builder.Build(cs.ctx))
			msS := cs.ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
			msS.Spec.MongoDB.Install = false
			cs.ctx.Set(constants.ContextSpec, msS)
			return cs
		},
		func() CaseStruct {
			cs := GenerateDefaultMongoTestCaseWrapper(
				"Only Monitoring Single Scheme",
				v1alpha1.SchemaSettings{
					SchemaType: v1alpha1.Single,
				},
			)
			cs.executor.SetExecutable(cs.builder.Build(cs.ctx))
			msS := cs.ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
			msS.Spec.MongoDB.Install = false
			cs.ctx.Set(constants.ContextSpec, msS)
			return cs
		},
		func() CaseStruct {
			cs := GenerateDefaultMongoTestCaseWrapper(
				"Single Schema settings validation",
				v1alpha1.SchemaSettings{
					SchemaType:      v1alpha1.Single,
					CnfReplicaSize:  1,
					DataReplicaSize: 2,
					ShardCount:      3,
					Sharded:         true,
				},
			)
			cs.executor.SetExecutable(cs.builder.Build(cs.ctx))
			msS := cs.ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
			cs.ctx.Set(constants.ContextSpec, msS)
			return cs
		},
		func() CaseStruct {
			cs := GenerateDefaultMongoTestCaseWrapper(
				"HA Schema settings empty values validation",
				v1alpha1.SchemaSettings{
					SchemaType: v1alpha1.HA,
				},
			)
			cs.executor.SetExecutable(cs.builder.Build(cs.ctx))
			msS := cs.ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
			cs.ctx.Set(constants.ContextSpec, msS)
			cs.ReadResultFunc = func(t *testing.T, err error) {
				if err == nil || err.Error() != utils.SchemaSettingsValidationZeros {
					t.Error("Expecting validation error...")
				}
			}
			return cs
		},
		func() CaseStruct {
			cs := GenerateDefaultMongoTestCaseWrapper(
				"HA Schema settings validation",
				v1alpha1.SchemaSettings{
					SchemaType:      v1alpha1.HA,
					CnfReplicaSize:  1,
					DataReplicaSize: 2,
					ShardCount:      3,
					Sharded:         true,
				},
			)
			cs.executor.SetExecutable(cs.builder.Build(cs.ctx))
			msS := cs.ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
			cs.ctx.Set(constants.ContextSpec, msS)
			cs.ReadResultFunc = func(t *testing.T, err error) {
				if err == nil || err.Error() != utils.SchemaSettingsValidationHAandArbiter {
					t.Error("Expecting validation error...")
				}
			}
			return cs
		},
		//TODO to services
		// func() CaseStruct {
		// 	cs := GenerateDefaultMongoTestCaseWrapper(
		// 		"HA Schema dynamic provisioner for backup",
		// 		v1alpha1.SchemaSettings{
		// 			SchemaType:      v1alpha1.HA,
		// 			CnfReplicaSize:  3,
		// 			DataReplicaSize: 3,
		// 			ShardCount:      3,
		// 			Sharded:         true,
		// 		},
		// 	)
		// 	cs.executor.SetExecutable(cs.builder.Build(cs.ctx))
		// 	msS := cs.ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
		// 	msS.Spec.Backup.Storage.NodeLabels = nil
		// 	msS.Spec.Backup.Storage.Volumes = nil
		// 	msS.Spec.Backup.Storage.StorageClasses = []string{"test"}
		// 	msS.Spec.Backup.AdditionalNodeLabels = nil
		// 	cs.ctx.Set(constants.ContextSpec, msS)
		// 	return cs
		// },
		func() CaseStruct {
			cs := GenerateDefaultMongoTestCaseWrapper(
				"Arbiter Schema settings validation",
				v1alpha1.SchemaSettings{
					SchemaType:      v1alpha1.Arbiter,
					CnfReplicaSize:  1,
					DataReplicaSize: 2,
					ShardCount:      3,
					Sharded:         true,
				},
			)
			cs.executor.SetExecutable(cs.builder.Build(cs.ctx))
			msS := cs.ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
			cs.ctx.Set(constants.ContextSpec, msS)
			cs.ReadResultFunc = func(t *testing.T, err error) {
				if err == nil || err.Error() != utils.SchemaSettingsValidationHAandArbiter {
					t.Error("Expecting validation error...")
				}
			}
			return cs
		},
		//TODO to services
		// func() CaseStruct {
		// 	cs := GenerateDefaultMongoTestCaseWrapper(
		// 		"EmptyDir check",
		// 		v1alpha1.SchemaSettings{
		// 			SchemaType: v1alpha1.Single,
		// 		},
		// 	)
		// 	msS := cs.ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
		// 	msS.Spec.Backup.Storage.EmptyDir = true
		// 	cs.executor.SetExecutable(cs.builder.Build(cs.ctx))
		// 	cs.ReadResultFunc = func(t *testing.T, err error) {
		// 		backup := &v12.Deployment{}
		// 		client := cs.ctx.Get(constants.ContextClient).(client.Client)
		// 		request := cs.ctx.Get(constants.ContextRequest).(reconcile.Request)
		// 		err = client.Get(context.TODO(),
		// 			types.NamespacedName{Name: utils.BackupDaemon, Namespace: request.Namespace}, backup)
		// 		if err != nil {
		// 			t.Error(err)
		// 		}

		// 		assert.True(t, backup.Spec.Template.Spec.Volumes[0].EmptyDir != nil)
		// 	}
		// 	return cs
		// },
		// func() CaseStruct {
		// 	cs := GenerateDefaultMongoTestCaseWrapper(
		// 		"Switchover",
		// 		v1alpha1.SchemaSettings{
		// 			SchemaType:      v1alpha1.DR,
		// 			CnfReplicaSize:  3,
		// 			DataReplicaSize: 3,
		// 			ShardCount:      1,
		// 			Sharded:         true,
		// 			OtherDomainName: "cluster.local",
		// 		},
		// 	)
		// 	msS := cs.ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
		// 	msS.Spec.DisasterRecovery.Mode = utils.ActiveMode
		// 	cs.executor.SetExecutable(cs.builder.Build(cs.ctx))
		// 	for key, elem := range cs.ctxToReplaceAfterServiceBuilt {
		// 		cs.ctx.Set(key, elem)
		// 	}
		// 	err := cs.executor.Execute(cs.ctx)
		// 	if err != nil {
		// 		panic(err)
		// 	}

		// 	msS.Spec.DisasterRecovery.Mode = utils.StandbyMode
		// 	cs.executor.SetExecutable(cs.drBuilder.Build(cs.ctx))
		// 	cs.ReadResultFunc = func(t *testing.T, err error) {
		// 		prom := &v12.Deployment{}
		// 		client := cs.ctx.Get(constants.ContextClient).(client.Client)
		// 		request := cs.ctx.Get(constants.ContextRequest).(reconcile.Request)
		// 		err = client.Get(context.TODO(),
		// 			types.NamespacedName{Name: utils.MongoPrometheusExporter, Namespace: request.Namespace}, prom)
		// 		if err != nil {
		// 			t.Error(err)
		// 		}

		// 		for _, env := range prom.Spec.Template.Spec.Containers[0].Env {
		// 			if env.Name == "EXPORT_MONGOS" {
		// 				assert.True(t, env.Value == strconv.FormatBool(false))
		// 			}
		// 		}

		// 	}
		// 	return cs
		// },
		func() CaseStruct {
			cs := GenerateDefaultMongoTestCaseWrapper(
				"Check replica scaling",
				v1alpha1.SchemaSettings{
					SchemaType:      v1alpha1.DR,
					CnfReplicaSize:  3,
					DataReplicaSize: 3,
					ShardCount:      1,
					Sharded:         true,
					OtherDomainName: "cluster.local",
				},
			)
			cs.executor.SetExecutable(cs.builder.Build(cs.ctx))
			for key, elem := range cs.ctxToReplaceAfterServiceBuilt {
				cs.ctx.Set(key, elem)
			}
			cs.executor.Execute(cs.ctx)
			cs.executor.SetExecutable(cs.drBuilder.Build(cs.ctx))
			return cs
		},
	}

	tests := []CaseStruct{}
	for _, tf := range testFuncs {
		tests = append(tests, tf())
	}
	defer gock.Off()

	for _, tt := range tests {
		if tt.name != "HA Schema dynamic provisioner for backup" {
			//continue
		}
		t.Run(tt.name, func(t *testing.T) {
			//tt.executor.SetExecutable(tt.builder.Build(tt.ctx))

			for key, elem := range tt.ctxToReplaceAfterServiceBuilt {
				tt.ctx.Set(key, elem)
			}

			err := tt.executor.Execute(tt.ctx)
			tt.ReadResultFunc(t, err)
		})
	}
}
