package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"go.uber.org/zap"
	v14 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type MongoHelper interface {
	GetMongosPod() (*v1.Pod, error)
	GetMongoPrimaryReplica(labelSelectors map[string]string) (*v1.Pod, error)
	GetMongoVersion() (string, error)
	MongoReplicaSetInit(nameKey string, nameKeyWithIndex string, replicaSetSize int, namespace string,
		domain string, configInit bool, memberOptionsFunc func(i int) string) error
	/* remove==== */
	RunOnMongos(arg string) (string, error)
	/* ===== */
	AddReplica(labels map[string]string, host string) error
	AddDRReplicas(labels map[string]string, replicasetHostnames []string) error
	ReconfigureRS(labels map[string]string, replicasetHostnames []string) error
	AddShards(domain string, dataReplicaSize, dataShards int) (err error)
	UpdateShardsInConfigDB(domain string, dataReplicaSize, dataShards int) (err error)
	UpdateConfigRSInDATARS(domain, drDomain string, cnfrReplicaSize, datarsSize, shardCount int) error
	GetClusterStatus(mode string, domain string, cnfReplicaSize int, dataReplicaSize int, shardCount int, sharded bool) (string, error)
	SetMongoCMD(cmd string)
	RestartMongos() error
	RestartConfigRS(cnfReplicaSize int, domain, mode string) error
	RemoveConfigRS(cnfReplicaSize int) error
	RemoveDataRS(dataReplicaSize, shardCount int) error
	RestartDataRS(datarsSize, shardCount int, domain, mode string) error
	ScaleMongos(replicas int) error
	ScaleDeployment(replicas int, microserviceName string) error
	UpdateDeployment(microserviceName string, updateFunc func(depl *v14.Deployment)) error
	ScaleDATARS(shardCount, replicas int) error
	ScaleCNFRS(replicas int) error
	RunOnShards(command string, shardCount int) ([]string, error)
	RunParallelOnShards(command string, shardCount int) error
	CheckUserExists(user string) bool
	CheckRoleExists(user string) bool
	UpdateRootPassword(authDB, username, oldPassword, newPassword string, inMonogs, inShards bool, shardsCound int) error
	CheckUserLogin(dockerImage, authDB, user, pass string) bool // we will get rid of docker image when drop 4.4 support
	CreateUser(authDB, user, pass, role string, force, inMonogs, inShards bool, shardsCound int) error
	CreateRole(authDB, role, privileges, roles string, inMonogs, inShards bool, shardsCound int) error
	SetFeatureCompatibilityVersion(sharded bool, shardCount int) error
	RunOnCnfrs(command string) (string, error)
	RunCompactCommand(shardsCount int, dbName string, collectionName string) error
	ExecuteFlushData(ctx core.ExecutionContext, shardsCount int) error
	Compact(shardsCount int, dbName string, collectionName string) error
	CompactAll(shardsCount int, dbName string) error
	GetRSStatus(labels map[string]string) string
	GetClusterRSStatus(cnfReplicaSize int, shardCount int, sharded bool) []string
	CheckFCV(shardsCount int) (bool, error)
	GetOplogSizes(ctx core.ExecutionContext, shardsCount int, creds *v1.Secret) (*OplogSizeReport, error)
	UpdateOplogSize(ctx core.ExecutionContext, oplogSize int64, report OplogSizeReport) error
}

var _ MongoHelper = &MongoUtilsHelperImpl{}

type MongoUtilsHelperImpl struct {
	KubernetesHelperImpl core.KubernetesHelper
	Client               client.Client
	KubeConfig           *rest.Config
	Logger               *zap.Logger
	Namespace            string
	Cmd                  string
	Tries                int
	RetryInterval        int
	WaitSeconds          int
	Sharded              bool
	Single               bool
}

type OplogSizeReport struct {
	Items []OplogSizeInfo
}

type OplogSizeInfo struct {
	ShardName  string
	ReplicaSet string
	PodName    string
	IsPrimary  bool
	MaxSizeMB  int64
	DomainName string
	Pod        *v1.Pod
}

type Member struct {
	Name       string `json:"name"`
	StateStr   string `json:"stateStr"`
	LagSeconds int    `json:"lagSeconds"`
}

func (r *MongoUtilsHelperImpl) UpdateOplogSize(ctx core.ExecutionContext, desiredOplogSize int64, report OplogSizeReport) error {
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	opLogResizeCmd := fmt.Sprintf(JsOpLogResize, desiredOplogSize)

	log.Sugar().Infof("CMD running for resize is : ", opLogResizeCmd)

	//Secondaries
	for _, item := range report.Items {
		if !item.IsPrimary && desiredOplogSize > item.MaxSizeMB {
			result, err := r.RunWithJSONResult(item.Pod, opLogResizeCmd)
			if err != nil {
				return err
			}

			label := map[string]string{
				Microservice: item.ShardName,
			}

			checkLagCmd := fmt.Sprintf(JsCheckMemberLag, item.DomainName)
			memberHealthy, err := r.RunOnPrimaryWithJSONResult(label, checkLagCmd)
			if err != nil {
				return err
			}

			log.Sugar().Infof("check lag cmd : ", checkLagCmd)

			log.Sugar().Infof("lagResult: ", memberHealthy)
			if memberHealthy == "false" {
				return fmt.Errorf("replication lag present")
			}

			log.Sugar().Infof("result of running on [%s] secondary %s", item.PodName, result)
		}
	}

	//Primaries
	for _, item := range report.Items {
		if item.IsPrimary && desiredOplogSize > item.MaxSizeMB {
			result, err := r.RunWithJSONResult(item.Pod, opLogResizeCmd)
			if err != nil {
				return err
			}

			checkHealthCmd := fmt.Sprintf(JsCheckMemberLag, item.DomainName)
			memberHealthy, err := r.RunWithJSONResult(item.Pod, checkHealthCmd)
			if err != nil {
				return err
			}

			log.Sugar().Infof("check health cmd : ", checkHealthCmd)

			log.Sugar().Infof("primary health Result: ", memberHealthy)
			if memberHealthy == "false" {
				return fmt.Errorf("replication lag present")
			}

			log.Sugar().Infof("result of running on [%s] primary %s", item.PodName, result)
		}
	}

	return nil
}

func (r *MongoUtilsHelperImpl) GetOplogSizes(ctx core.ExecutionContext, shardsCount int, creds *v1.Secret) (*OplogSizeReport, error) {
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)

	report := &OplogSizeReport{
		Items: make([]OplogSizeInfo, 0),
	}

	for i := 0; i < shardsCount; i++ {
		dKey := fmt.Sprintf(DataNameKey, i+1)

		label := map[string]string{
			Microservice: dKey,
		}

		// list ALL pods in shard (primary + secondaries)
		podList, err := checkListPodsResult(
			r.KubernetesHelperImpl.ListPods(r.Namespace, label),
		)
		if err != nil {
			return nil, err
		}

		for _, pod := range podList.Items {
			opLogCmd := fmt.Sprintf(JsOpLogSize, "local")
			r.Cmd = fmt.Sprintf(
				MongoCMDAuthTemplate,
				MongoBinary(spec.Spec.MongoDB.DockerImage),
				string(creds.Data[Username]),
				string(creds.Data[Password]),
				spec.Spec.AuthDb,
			)

			log.Sugar().Infof("cmd being passed: %s", r.Cmd)

			output, err := r.RunOnMongoPod(&pod, opLogCmd)
			if err != nil {
				return nil, fmt.Errorf("failed oplog fetch for pod %s: %w", pod.Name, err)
			}
			log.Sugar().Infof("nAME : %s", pod.Name)
			log.Sugar().Infof("output is : %s", output)

			sizeBytes, err := strconv.ParseInt(strings.TrimSpace(output), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("parse oplog failed for pod %s: %w", pod.Name, err)
			}

			// detect primary (optional safety: do not assume helper is already used)
			isPrimary := false
			if primary, err := r.GetMongoPrimaryReplica(label); err == nil {
				isPrimary = primary.Name == pod.Name
			}

			report.Items = append(report.Items, OplogSizeInfo{
				ShardName:  dKey,
				ReplicaSet: dKey,
				DomainName: fmt.Sprintf("%s.%s.%s:27017", pod.Name, dKey, fmt.Sprintf("%s.svc.%s", request.Namespace, spec.Spec.SchemaSettings.ThisDomainName)),
				PodName:    pod.Name,
				IsPrimary:  isPrimary,
				MaxSizeMB:  sizeBytes / (1024 * 1024),
				Pod:        &pod,
			})
		}
	}

	return report, nil
}

func (r *MongoUtilsHelperImpl) RunOnShards(command string, shardCount int) ([]string, error) {
	result := []string{}
	for i := 0; i < shardCount; i++ {
		dKey := fmt.Sprintf(DataNameKey, i+1)

		res, err := r.RunOnPrimary(
			map[string]string{
				Microservice: dKey,
			},
			command,
		)
		if err != nil {
			return result, err
		}
		res = strings.TrimSpace(res)
		result = append(result, res)
	}
	return result, nil
}

func (r *MongoUtilsHelperImpl) RunCompactCommand(shardsCount int, dbName string, collectionName string) error {
	for i := 0; i < shardsCount; i++ {
		dKey := fmt.Sprintf(DataNameKey, i+1)
		label := map[string]string{
			Microservice: dKey,
		}
		primaryReplicaNode, err := r.GetMongoPrimaryReplica(label)
		if err != nil {
			return err
		}
		list, err := checkListPodsResult(r.KubernetesHelperImpl.ListPods(r.Namespace, label))
		if err != nil {
			return err
		}

		for _, pod := range list.Items {
			if pod.Name != primaryReplicaNode.Name {
				_, err = r.RunOnMongoPod(&pod, fmt.Sprintf("db.getSiblingDB('%s').runCommand({compact: '%s'})", dbName, collectionName))
			}
		}
		_, err = r.RunOnMongoPod(primaryReplicaNode, "rs.stepDown()")
		if err != nil {
			return err
		}
		for {
			primaryReplicaNodeSecond, err := r.GetMongoPrimaryReplica(label)
			if err != nil {
				return err
			}
			if primaryReplicaNode.Name != primaryReplicaNodeSecond.Name {
				break
			}
			time.Sleep(5 * time.Second)
		}
		_, err = r.RunOnMongoPod(primaryReplicaNode, fmt.Sprintf("db.getSiblingDB('%s').runCommand({compact: '%s'})", dbName, collectionName))

	}
	return nil
}

func (r *MongoUtilsHelperImpl) RunCompactCommandForALL(shardsCount int, dbName string) error {
	for i := 0; i < shardsCount; i++ {
		dKey := fmt.Sprintf(DataNameKey, i+1)
		label := map[string]string{
			Microservice: dKey,
		}
		primaryReplicaNode, err := r.GetMongoPrimaryReplica(label)
		if err != nil {
			return err
		}
		list, err := checkListPodsResult(r.KubernetesHelperImpl.ListPods(r.Namespace, label))
		if err != nil {
			return err
		}

		for _, pod := range list.Items {
			if pod.Name != primaryReplicaNode.Name {
				_, err := r.RunOnMongoPod(
					primaryReplicaNode, "rs.secondaryOk()")
				if err != nil {
					return err
				}
				_, err = r.RunOnMongoPod(&pod, fmt.Sprintf("db.getSiblingDB(%s).getCollectionNames().forEach(function (collectionName) { db.runCommand({ compact: collectionName }) })", dbName))
			}
		}
		_, err = r.RunOnMongoPod(primaryReplicaNode, fmt.Sprintf("db.getSiblingDB(%s).getCollectionNames().forEach(function (collectionName) { db.runCommand({ compact: collectionName }) })", dbName))

	}
	return nil
}

func (r *MongoUtilsHelperImpl) RunParallelOnShards(command string, shardCount int) error {
	var wg sync.WaitGroup

	result := make(chan error, shardCount)

	for i := 0; i < shardCount; i++ {
		wg.Add(1)

		i := i

		go func() {
			defer wg.Done()
			dKey := fmt.Sprintf(DataNameKey, i+1)
			_, err := r.RunOnPrimary(
				map[string]string{
					Microservice: dKey,
				},
				command,
			)
			result <- err
		}()
	}

	wg.Wait()

	close(result)

	for err := range result {
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *MongoUtilsHelperImpl) RunOnCnfrs(command string) (string, error) {

	result, err := r.RunOnPrimary(
		map[string]string{
			Microservice: CnfNameKey,
		},
		command,
	)

	return result, err
}

func (r *MongoUtilsHelperImpl) CheckUserLogin(dockerImage, authDB, user, pass string) bool {
	pod, err := r.GetMongosPod()
	if err != nil {
		return false
	}
	_, err = r.KubernetesHelperImpl.ExecRemote(
		nil,
		r.KubeConfig,
		pod.Name,
		pod.Namespace,
		pod.Spec.Containers[0].Name,
		BashCommand,
		[]string{fmt.Sprintf(MongoCheckUserLogin, MongoBinary(dockerImage), user, pass, authDB)})

	return err == nil

}

// TODO createUserOrRole
func (r *MongoUtilsHelperImpl) CreateRole(authDB, role, privileges, roles string, inMonogs, inShards bool, shardsCound int) error {
	command := fmt.Sprintf(CreateORUpdateRolePattern, authDB, role, privileges, roles)
	if inMonogs {
		_, err := r.RunOnMongos(command)
		if err != nil {
			return err
		}
	}

	if inShards {
		_, err := r.RunOnShards(command, shardsCound)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *MongoUtilsHelperImpl) CreateUser(authDB, user, pass, role string, force, inMonogs, inShards bool, shardsCound int) error {
	command := fmt.Sprintf(CreateORUpdateUserPattern, authDB, user, pass, role)
	if force {
		command = CreateUserForMongoReplicaCommand(authDB, user, pass, role)
	}
	if inMonogs {
		_, err := r.RunOnMongos(command)
		if err != nil {
			return err
		}
	}

	if inShards {
		_, err := r.RunOnShards(command, shardsCound)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *MongoUtilsHelperImpl) UpdateRootPassword(authDB, username, oldPassword, newPassword string, inMonogs, inShards bool, shardsCound int) error {
	// "db.getSiblingDB('%s').UpdateUser('%s', {pwd: '%s'})"
	command := fmt.Sprintf(UpdateRootPassword, authDB, username, newPassword)

	if inMonogs {
		_, err := r.RunOnMongos(command)
		if err != nil {
			return err
		}
	}

	if inShards {
		_, err := r.RunOnShards(command, shardsCound)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *MongoUtilsHelperImpl) checkUserOrRoleExists(commandPattern, roleOrUser string) bool {

	var res string
	var err error
	if r.Sharded || r.Single {
		res, err = r.RunOnMongos(
			fmt.Sprintf(commandPattern, roleOrUser),
		)
	} else {
		res, err = r.RunOnPrimary(map[string]string{Microservice: fmt.Sprintf(DataNameKey, 1)},
			fmt.Sprintf(commandPattern, roleOrUser),
		)
	}

	if err != nil {
		r.Logger.Warn(fmt.Sprintf("A problem during %s roleOrUser check: %v . If this happens during clean install - ignore this message", roleOrUser, err.Error()))
		return false
	}

	if strings.Contains(strings.ReplaceAll(res, "\n", ""), roleOrUser) {
		r.Logger.Info("Mongos " + roleOrUser + " roleOrUser is already created")
		return true
	} else {
		return false
	}
}

func (r *MongoUtilsHelperImpl) CheckUserExists(user string) bool {
	command := "dbUsers = db.getUsers();" +
		"found = dbUsers.users?dbUsers.users.find((v)=>v.user=='%s').user:dbUsers.find((v)=>v.user=='%s').user;"
	return r.checkUserOrRoleExists(command, user)
}

func (r *MongoUtilsHelperImpl) CheckRoleExists(role string) bool {
	command := "dbRoles = db.getRoles();" +
		"found = dbRoles.roles?dbRoles.roles.find((v)=>v.role=='%s').role:dbRoles.find((v)=>v.role=='%s').role;"
	return r.checkUserOrRoleExists(command, role)
}

func (r *MongoUtilsHelperImpl) GetMongosPod() (*v1.Pod, error) {
	list, err := checkListPodsResult(r.KubernetesHelperImpl.ListPods(r.Namespace, map[string]string{
		Name: Mongos,
	}))
	if err != nil {
		return nil, err
	}

	return &list.Items[0], nil
}

func (r *MongoUtilsHelperImpl) GetMongoPrimaryReplica(labelSelectors map[string]string) (*v1.Pod, error) {
	list, err := checkListPodsResult(r.KubernetesHelperImpl.ListPods(r.Namespace, labelSelectors))
	if err != nil {
		return nil, err
	} else {
		response, err := r.RunOnMongoPod(
			&list.Items[0],
			PrimaryReplicaCommand)

		if err != nil {
			return nil, err
		}

		response = strings.TrimSuffix(response, "\n")
		responseSlice := strings.SplitAfter(response, "\n")
		if len(responseSlice) > 1 {
			response = responseSlice[1]
		}

		if response != "" {
			var masterPod *v1.Pod
			for _, pod := range list.Items {
				if pod.Name == response {
					masterPod = &pod
					break
				}
			}
			if masterPod != nil {
				return masterPod, nil
			}
		}

		return nil, &core.ExecutionError{Msg: "Not found primary for these selectors: " + fmt.Sprint(labelSelectors)}
	}
}

func (r *MongoUtilsHelperImpl) GetMongoVersion() (string, error) {
	if r.Single {
		return r.RunOnMongos(versionCmd)
	}

	return r.RunOnPrimary(map[string]string{Microservice: fmt.Sprintf(DataNameKey, 1)}, versionCmd)
}

func (r *MongoUtilsHelperImpl) MongoReplicaSetInit(nameKey string, nameKeyWithIndex string, replicaSetSize int, namespace string,
	domain string, configInit bool, memberOptionsFunc func(i int) string) error {
	// Build all cnfrs pods string according to cnfReplicaSize
	fullDomain := fmt.Sprintf("%s.svc.%s", namespace, domain)
	members := ""
	splitter := ""
	for i := 0; i < replicaSetSize; i++ {
		podNameWithIndex := fmt.Sprintf(nameKeyWithIndex, i)
		podFullName := fmt.Sprintf(StatefulSetPodNameTemplate, podNameWithIndex)

		options := memberOptionsFunc(i)

		members += fmt.Sprintf("%s{_id:%v, host:'%s.%s.%s:27017'%s}", splitter, i, podFullName, nameKey, fullDomain, options)
		splitter = ","
	}

	//Select first cnfrs pod
	podNameWithIndex := fmt.Sprintf(nameKeyWithIndex, 0)
	podFullName := fmt.Sprintf(StatefulSetPodNameTemplate, podNameWithIndex)

	//Prepare initialization string
	configShard := ""
	if configInit {
		configShard = "configsvr:true,"
	}
	jsCommand := fmt.Sprintf(
		"rs.initiate({_id:'%s', %s members:[%s]})",
		nameKey, configShard, members)

	// Initiate with all pods
	// var result OutputResult
	// Check if host is reachable
	pod := &v1.Pod{
		ObjectMeta: v12.ObjectMeta{
			Name:      podFullName,
			Namespace: namespace,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: podNameWithIndex,
				},
			},
		},
	}

	executionError := r.runAndRetry(func() error {
		return r.CheckOutputResult(r.RunWithJSONResult(pod, jsCommand))
	})

	if executionError != nil {
		return executionError
	}

	err := wait.PollImmediate(2*time.Second, time.Second*time.Duration(r.WaitSeconds), func() (done bool, err error) {
		command := "rs.status().members.filter(function(x){return x.state==1}).length"
		output, err := r.RunOnMongoPod(
			pod,
			command)

		// Cnf with "1" state is not found which means leader is not elected
		output = strings.TrimSpace(output)
		if output == "" {
			// No output means likely no primary yet — keep retrying
			r.Logger.Info(fmt.Sprintf("Empty output from command on pod %s, primary not elected yet\n", pod.Name))
			return false, nil
		}
		output = strings.TrimSuffix(output, "\n") // Workaround
		i, err := strconv.Atoi(output)
		if err != nil {
			out := strings.SplitAfter(output, "\n")
			if len(out) > 1 {
				i, err = strconv.Atoi(out[1])
			}
		}
		return i == 1, err
	})

	if err != nil {
		return &core.ExecutionError{Msg: "Error happened on " + nameKey + " primary election. Error: " + err.Error()}
	}

	return nil
}

func (r *MongoUtilsHelperImpl) RunWithJSONResult(masterPod *v1.Pod, arg string) (string, error) {
	arg = fmt.Sprintf("JSON.stringify(%s);", arg)
	return r.RunOnMongoPod(masterPod, arg)
}

func (r *MongoUtilsHelperImpl) RunOnPrimaryWithJSONResult(label map[string]string, arg string) (string, error) {
	arg = fmt.Sprintf("JSON.stringify(%s);", arg)
	return r.RunOnPrimary(label, arg)
}

func (r *MongoUtilsHelperImpl) RunOnMongoPod(masterPod *v1.Pod, arg string) (string, error) {
	resp, err := r.KubernetesHelperImpl.ExecRemote(
		nil,
		r.KubeConfig,
		masterPod.Name,
		masterPod.Namespace,
		masterPod.Spec.Containers[0].Name,
		BashCommand,
		[]string{fmt.Sprintf("%s --eval \"%s\"", r.Cmd, arg)})

	const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
	var re = regexp.MustCompile(ansi)
	if err == nil {
		resp = re.ReplaceAllString(resp, "")
		resp = strings.Replace(resp, MongoshBanner7, "", 1)
		resp = strings.Replace(resp, MongoshBanner8, "", 1)
	}

	return resp, err
}

func (r *MongoUtilsHelperImpl) AddReplica(labels map[string]string, host string) error {
	arg := fmt.Sprintf("rs.add( { host: '%s'} )", host)
	err := r.CheckOutputResult(r.RunOnPrimaryWithJSONResult(labels, arg))

	if err != nil {
		return err
	}

	return nil
}

func (r *MongoUtilsHelperImpl) AddHiddenReplica(labels map[string]string, host string) error {
	arg := fmt.Sprintf("rs.add( { host: '%s', priority: 0, votes: 0 } )", host)
	err := r.CheckOutputResult(r.RunOnPrimaryWithJSONResult(labels, arg))

	if err != nil {
		return err
	}

	return nil
}

func (r *MongoUtilsHelperImpl) AddDRReplicas(labels map[string]string, replicasetHostnames []string) error {
	for _, host := range replicasetHostnames {
		err := r.runAndRetry(func() error {
			return r.AddHiddenReplica(labels, host)
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *MongoUtilsHelperImpl) ReconfigureRS(labels map[string]string, replicasetHostnames []string) error {
	list, err := checkListPodsResult(r.KubernetesHelperImpl.ListPods(r.Namespace, labels))

	if err != nil {
		return nil
	}

	QuoteReplicaSet(replicasetHostnames)

	arg := fmt.Sprintf(JsTemplate, strings.Join(replicasetHostnames, ","))
	r.runAndRetry(func() error {
		return r.CheckOutputResult(r.RunOnMongoPod(&list.Items[0], arg))
	})
	return err
}

func (r *MongoUtilsHelperImpl) AddOrUpdateShards(domain string, dataReplicaSize, dataShards int, commandFormatFunc func(shardName, replicas string) string,
	commandFunc func(c string) error) (err error) {
	for i := 0; i < dataShards; i++ {
		dKey := fmt.Sprintf(DataNameKey, i+1)
		dataReplicas := GetDATAReplicaSetHostName(dataReplicaSize, i, domain, r.Namespace)

		js := commandFormatFunc(dKey, strings.Join(dataReplicas, ","))
		err = r.runAndRetry(func() error {
			return commandFunc(js)
		})
	}
	return
}

func (r *MongoUtilsHelperImpl) AddShards(domain string, dataReplicaSize, dataShards int) (err error) {
	return r.AddOrUpdateShards(domain, dataReplicaSize, dataShards,
		func(shardName, replicas string) string {
			return fmt.Sprintf("JSON.stringify(sh.addShard('%s/%s'))", shardName, replicas)
		},
		func(command string) error {
			return r.CheckOutputResult(r.RunOnMongos(command))
		})
}

func (r *MongoUtilsHelperImpl) UpdateShardsInConfigDB(domain string, dataReplicaSize, dataShards int) (err error) {
	return r.AddOrUpdateShards(domain, dataReplicaSize, dataShards,
		func(shardName, replicas string) string {
			return fmt.Sprintf(shardHostsTemplate, shardName, shardName, replicas)
		},
		func(command string) error {
			_, err := r.RunOnPrimary(map[string]string{Microservice: CnfNameKey}, command)
			return err
		})
}

func (r *MongoUtilsHelperImpl) runAndRetry(runFunc func() error) (err error) {
	for i := 0; i < r.Tries; i++ {
		err = runFunc()
		if err != nil {
			time.Sleep(time.Duration(r.RetryInterval) * time.Second)
		} else {
			break
		}
	}

	return
}

func (r *MongoUtilsHelperImpl) UpdateConfigRSInDATARS(domain, drDomain string, cnfrReplicaSize, datarsSize, shardCount int) error {
	configRS := GetDRCNFReplicaSetHostNames(cnfrReplicaSize, domain, drDomain, r.Namespace)
	for s := 0; s < shardCount; s++ {
		dKey := fmt.Sprintf(DataNameKey, s+1)
		shardJs := fmt.Sprintf(shardTemplate, dKey, strings.Join(configRS, ","))
		res, updateErr := r.RunOnPrimary(map[string]string{Microservice: dKey}, shardJs)
		r.Logger.Debug(fmt.Sprintf("Update shard with command %s return with %s", shardJs, res))
		if updateErr != nil {
			return updateErr
		}
	}
	return nil
}

func (r *MongoUtilsHelperImpl) GetReplicaSetStatus(labels map[string]string, replicasetHostnames []string, mode string) (string, error) {
	list, err := checkListPodsResult(r.KubernetesHelperImpl.ListPods(r.Namespace, labels))

	if err != nil {
		return "", nil
	}

	QuoteReplicaSet(replicasetHostnames)

	replicasStatesCommand := fmt.Sprintf(JsReplicasWithState, strings.Join(replicasetHostnames, ","))
	replicasJson, err := r.RunWithJSONResult(&list.Items[0], replicasStatesCommand)

	if err != nil {
		return "", err
	}

	var replicas []string

	jsonErr := json.Unmarshal([]byte(replicasJson), &replicas)

	if jsonErr != nil {
		return "", jsonErr
	}

	r.Logger.Debug(fmt.Sprintf("rs.status() is %v", replicas))

	var secondaryReplicas int
	var primaryReplicas int
	for _, replica := range replicas {
		if replica == "SECONDARY" {
			secondaryReplicas++
		} else if replica == "PRIMARY" {
			primaryReplicas++
		}
	}

	if mode == ActiveMode {
		if primaryReplicas == 0 {
			return Down, nil
		} else if secondaryReplicas == len(replicasetHostnames)-1 {
			return Up, nil
		}
	} else {
		if secondaryReplicas == 0 {
			return Down, nil
			// In case if current active was reconfigured to standby, primary will be elected on current (not yet activated standby)
		} else if secondaryReplicas == len(replicasetHostnames) ||
			(secondaryReplicas == len(replicasetHostnames)-1 && primaryReplicas == 1) {
			return Up, nil
		}
	}
	return Degraded, nil
}

func (r *MongoUtilsHelperImpl) GetClusterStatus(mode string, domain string, cnfReplicaSize int, dataReplicaSize int, shardCount int, sharded bool) (string, error) {

	type rsstatus struct {
		indx   int
		result string
		err    error
	}

	dataStatusChannel := make(chan rsstatus, shardCount)
	cnfrsStatusChannel := make(chan rsstatus, 1)

	var err error

	var wg sync.WaitGroup

	for i := 0; i < shardCount; i++ {
		wg.Add(1)

		i := i
		go func() {
			defer wg.Done()
			rsName := GetDATAReplicaSetHostName(dataReplicaSize, i, domain, r.Namespace)
			serviceName := fmt.Sprintf(DataNameKey, i+1)
			dataStatus, err := r.GetReplicaSetStatus(map[string]string{Microservice: serviceName}, rsName, mode)
			r.Logger.Debug(fmt.Sprintf("Replicaset %s status is %s", rsName, dataStatus))

			dataStatusChannel <- rsstatus{i, dataStatus, err}
		}()
	}

	if cnfReplicaSize > 0 && sharded {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cnfStatus, err := r.GetReplicaSetStatus(map[string]string{Microservice: CnfNameKey}, GetCNFReplicaSetHostNames(cnfReplicaSize, domain, r.Namespace), mode)

			cnfrsStatusChannel <- rsstatus{0, cnfStatus, err}
		}()
	}

	wg.Wait()

	close(dataStatusChannel)
	close(cnfrsStatusChannel)

	dataStatus := make([]string, shardCount)
	for status := range dataStatusChannel {
		if status.err != nil {
			return Down, err
		}
		dataStatus[status.indx] = status.result
	}

	var cnfStatus string
	if cnfReplicaSize > 0 && sharded {
		for status := range cnfrsStatusChannel {
			if status.err != nil {
				return Down, err
			}
			cnfStatus = status.result
		}
	}

	if sharded {
		dataDown := 0
		for d := 0; d < shardCount; d++ {
			if dataStatus[d] == Down {
				dataDown++
			}
		}

		if mode == ActiveMode && sharded {
			pod, err := r.GetMongosPod()
			if err != nil {
				return Down, err
			}

			if pod.Status.Phase != v1.PodRunning {
				return Down, nil
			}
		}

		if cnfStatus == Down || dataDown == shardCount {
			return Down, nil
		}

		if cnfStatus == Degraded || (cnfStatus == Up && dataDown > 0) {
			return Degraded, nil
		} else if cnfStatus == Up && dataDown == 0 {
			return Up, nil
		}

		return Down, nil
	} else {
		return dataStatus[0], nil
	}
}

func (r *MongoUtilsHelperImpl) GetClusterRSStatus(cnfReplicaSize int, shardCount int, sharded bool) []string {
	clsStatus := make([]string, shardCount)

	for i := 0; i < shardCount; i++ {
		serviceName := fmt.Sprintf(DataNameKey, i+1)
		clsStatus[i] = r.GetRSStatus(map[string]string{Microservice: serviceName})
	}

	if sharded {
		if cnfReplicaSize > 0 && sharded {
			clsStatus = append(clsStatus, r.GetRSStatus(map[string]string{Microservice: CnfNameKey}))
		}
	}
	return clsStatus
}

func (r *MongoUtilsHelperImpl) RunOnPrimary(labels map[string]string, arg string) (res string, err error) {
	r.runAndRetry(func() error {
		masterPod, errr := r.GetMongoPrimaryReplica(labels)

		if errr != nil {
			return errr
		} else {
			res, err = r.RunOnMongoPod(
				masterPod,
				arg)

			return err
		}
	})

	return
}

func (r *MongoUtilsHelperImpl) RunOnMongos(arg string) (string, error) {
	pod, err := r.GetMongosPod()
	if err != nil {
		return "", err
	}
	res, errr := r.RunOnMongoPod(
		pod,
		arg)

	return res, errr
}

func (r *MongoUtilsHelperImpl) ScaleMongos(replicas int) error {
	mongos := &v1.ReplicationController{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: Mongos, Namespace: r.Namespace}, mongos)
	if err != nil {
		return err
	}
	return r.KubernetesHelperImpl.ScaleReplicationController(mongos, replicas, r.WaitSeconds)
}

func (r *MongoUtilsHelperImpl) CheckOutputResult(output string, err error) error {
	if err != nil {
		return err
	}

	var result OutputResult

	jsonErr := json.Unmarshal([]byte(output), &result)

	if jsonErr != nil {
		outputString := strings.SplitAfter(output, "\n")
		if len(outputString) > 1 {
			jsonErr = json.Unmarshal([]byte(outputString[1]), &result)
		}
	}

	if jsonErr != nil {
		return &core.ExecutionError{
			Msg: fmt.Sprintf("Unmarshaling error. Error: %s.", jsonErr.Error()),
		}
	} else if result.Ok != 1 {
		return &core.ExecutionError{
			Msg: fmt.Sprintf("raw output = %s, result.code = %d, result.errmgs = %s", output, result.Code, result.ErrMsg),
		}
	}

	return nil
}

func (r *MongoUtilsHelperImpl) SetMongoCMD(cmd string) {
	r.Cmd = cmd
}

func (r *MongoUtilsHelperImpl) RestartMongos() error {
	pod, err := r.GetMongosPod()
	if err != nil {
		return fmt.Errorf("error while receiving mongos pod. Error: %v", err)
	}

	err = r.KubernetesHelperImpl.RestartPod(pod, r.Namespace, r.WaitSeconds)
	if err != nil {
		return fmt.Errorf("error while restarting mongos pod. Error: %v", err)
	}

	return nil
}

func (r *MongoUtilsHelperImpl) RestartConfigRS(cnfReplicaSize int, domain, mode string) error {
	list, err := checkListPodsResult(r.KubernetesHelperImpl.ListPods(r.Namespace, map[string]string{Microservice: CnfNameKey}))
	if err != nil {
		return fmt.Errorf("error while receiving cfrs pod. Error: %v", err)
	}

	for i := 0; i < len(list.Items); i++ {
		err = r.KubernetesHelperImpl.RestartPod(&list.Items[i], r.Namespace, r.WaitSeconds)
		if err != nil {
			return fmt.Errorf("error while restarting cnfrs %d pod. Error: %v", i, err)
		}

		if i == 0 {
			err := wait.PollImmediate(3*time.Second, time.Second*time.Duration(r.WaitSeconds), func() (done bool, err error) {
				cnfStatus, statusErr := r.GetReplicaSetStatus(map[string]string{Microservice: CnfNameKey}, GetCNFReplicaSetHostNames(cnfReplicaSize, domain, r.Namespace), mode)
				if statusErr != nil || cnfStatus != Up {
					return false, nil
				}
				return true, nil
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *MongoUtilsHelperImpl) RemoveConfigRS(cnfReplicaSize int) error {
	for i := 0; i < cnfReplicaSize; i++ {
		err := r.KubernetesHelperImpl.DeleteStatefulsetAndPods(fmt.Sprintf("%s%d", CnfNameKey, i), r.Namespace, r.WaitSeconds)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *MongoUtilsHelperImpl) RemoveDataRS(dataReplicaSize, shardCount int) error {
	for i := 1; i < shardCount+1; i++ {
		for j := 0; j < dataReplicaSize; j++ {
			err := r.KubernetesHelperImpl.DeleteStatefulsetAndPods(fmt.Sprintf(fmt.Sprintf(DataNameKey, "%d%d"), i, j), r.Namespace, r.WaitSeconds)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *MongoUtilsHelperImpl) RestartDataRS(datarsSize, shardCount int, domain, mode string) error {
	for s := 0; s < shardCount; s++ {
		serviceName := fmt.Sprintf(DataNameKey, s+1)
		list, err := checkListPodsResult(r.KubernetesHelperImpl.ListPods(r.Namespace, map[string]string{Microservice: serviceName}))

		if err != nil {
			return err
		}

		for i := 0; i < len(list.Items); i++ {
			err = r.KubernetesHelperImpl.RestartPod(&list.Items[i], r.Namespace, r.WaitSeconds)
			if err != nil {
				return err
			}

			if i == 0 {
				err := wait.PollImmediate(3*time.Second, time.Second*time.Duration(r.WaitSeconds), func() (done bool, err error) {
					dataStatus, statusErr := r.GetReplicaSetStatus(map[string]string{Microservice: serviceName}, GetDATAReplicaSetHostName(datarsSize, s, domain, r.Namespace), mode)
					if statusErr != nil || dataStatus != Up {
						return false, nil
					}
					return true, nil
				})
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func checkListPodsResult(list *v1.PodList, err error) (*v1.PodList, error) {
	if err != nil {
		return nil, err
	}
	if len(list.Items) == 0 {
		return nil, fmt.Errorf("pod list is empty")
	}
	return list, nil
}

func (r *MongoUtilsHelperImpl) ScaleDeployment(replicas int, microserviceName string) error {
	dl := &v14.DeploymentList{}
	labels := map[string]string{
		Microservice: microserviceName,
	}
	err := r.KubernetesHelperImpl.ListRuntimeObjectsByLabels(dl, r.Namespace, labels)
	if err != nil {
		return err
	}
	return r.KubernetesHelperImpl.ScaleDeployment(&dl.Items[0], replicas, r.WaitSeconds)
}

func (r *MongoUtilsHelperImpl) UpdateDeployment(microserviceName string, updateFunc func(depl *v14.Deployment)) error {
	dl := &v14.DeploymentList{}
	labels := map[string]string{
		Microservice: microserviceName,
	}

	err := r.KubernetesHelperImpl.ListRuntimeObjectsByLabels(dl, r.Namespace, labels)
	if err != nil {
		return err
	}
	if len(dl.Items) == 0 {
		return fmt.Errorf("No deployment found by labels microservice: %s", microserviceName)
	}

	updateFunc(&dl.Items[0])

	return r.Client.Update(context.TODO(), &dl.Items[0], &client.UpdateOptions{})
}

func (r *MongoUtilsHelperImpl) ScaleCNFRS(replicas int) error {
	ss := &v14.StatefulSetList{}
	labels := map[string]string{
		Microservice: CnfNameKey,
	}
	err := r.KubernetesHelperImpl.ListRuntimeObjectsByLabels(ss, r.Namespace, labels)
	if err != nil {
		return err
	}
	for _, item := range ss.Items {
		err = r.KubernetesHelperImpl.ScaleStatefulset(&item, replicas, r.WaitSeconds)
		if err != nil {
			return nil
		}
	}
	return nil
}

func (r *MongoUtilsHelperImpl) ScaleDATARS(shardCount, replicas int) error {
	for i := 1; i <= shardCount; i++ {
		ss := &v14.StatefulSetList{}
		labels := map[string]string{
			Microservice: fmt.Sprintf(DataNameKey, i),
		}
		err := r.KubernetesHelperImpl.ListRuntimeObjectsByLabels(ss, r.Namespace, labels)
		if err != nil {
			return err
		}
		for _, item := range ss.Items {
			err = r.KubernetesHelperImpl.ScaleStatefulset(&item, replicas, r.WaitSeconds)
			if err != nil {
				return nil
			}
		}
	}
	return nil
}

func (r *MongoUtilsHelperImpl) SetFeatureCompatibilityVersion(sharded bool, dataRSSize int) error {
	r.Logger.Debug("Start setting new featureCompatibilityVersion value")

	re := regexp.MustCompile(`[0-9]\.[0-9]`)

	rawVersion, err := r.GetMongoVersion()
	if err != nil {
		r.Logger.Error(fmt.Sprintf("cannot get mongo version. error: %s", err.Error()))
		return err
	}

	r.Logger.Debug(fmt.Sprintf("db version is LOOK HERE: %s", rawVersion))

	version := re.FindString(rawVersion)

	r.Logger.Debug(fmt.Sprintf("minor version is LOOK HERE %s", version))

	if version == "" {
		return errors.New("cannot get major and minor version from mongo version")
	}

	command := fmt.Sprintf(GetSetFeatureCompatibilityVersion(version), version)

	if r.Single {
		_, err = r.RunOnMongos(command)
		if err != nil {
			return err
		}
	} else {
		if r.Sharded {
			_, err = r.RunOnCnfrs(command)
			if err != nil {
				return err
			}
		}

		_, err = r.RunOnShards(command, dataRSSize)
	}

	return err
}

func (r *MongoUtilsHelperImpl) ExecuteFlushData(ctx core.ExecutionContext, shardsCount int) error {
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
	log.Debug("db.adminCommand( { fsync: 1, lock: false } ) command execution is started")

	_, err := r.RunOnShards(fmt.Sprintf("db.adminCommand( { fsync: 1, lock: false } )"), shardsCount)
	if err != nil {
		return err
	}
	return nil
}

func (r *MongoUtilsHelperImpl) Compact(shardsCount int, dbName string, collectionName string) error {
	err := r.RunCompactCommand(shardsCount, dbName, collectionName)
	if err != nil {
		return err
	}
	return nil
}

func (r *MongoUtilsHelperImpl) CompactAll(shardsCount int, dbName string) error {
	err := r.RunCompactCommandForALL(shardsCount, dbName)
	if err != nil {
		return err
	}
	return nil
}

func (r *MongoUtilsHelperImpl) GetRSStatus(labels map[string]string) string {
	list, err := checkListPodsResult(r.KubernetesHelperImpl.ListPods(r.Namespace, labels))

	if err != nil {
		return "Error while fetching pod list"
	}
	replicas, err := r.RunWithJSONResult(&list.Items[0], JsReplicasWithStateAndName)

	if err != nil {
		return fmt.Sprintf("Error while running command %s", JsReplicasWithStateAndName)
	}
	return replicas
}

func (r *MongoUtilsHelperImpl) CheckFCV(shardCount int) (bool, error) {
	cnfrsFCVs, err := r.RunOnCnfrs(featureCompatibilityVersionCommand)
	cnfrsFCVs = strings.TrimSpace(cnfrsFCVs)
	if err != nil {
		return true, err
	}
	shardsFCVs, err := r.RunOnShards(featureCompatibilityVersionCommand, shardCount)
	if err != nil {
		return true, err
	}
	for _, element := range shardsFCVs {
		if element != cnfrsFCVs {
			return true, nil
		}
	}
	return false, nil
}
