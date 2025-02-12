package utils

// import (
// 	"fmt"
// 	"os/exec"
// 	"testing"

// 	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core/mocks"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// 	"github.com/stretchr/testify/suite"
// 	"go.uber.org/zap"
// 	v1 "k8s.io/api/core/v1"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/client-go/rest"
// )

// type TestSuite struct {
// 	suite.Suite
// }

// func TestExampleTestSuite(t *testing.T) {
// 	suite.Run(t, new(TestSuite))
// }

// func (suite *TestSuite) SetupSuite() {
// 	_, err := exec.Command("/usr/bin/podman", []string{"run", "--name", "mongodb", "-p", "27017:27017", "artifactorycn.netcracker.com:17064/mongo:4.2.23"}...).Output()
// 	if err != nil {
// 		suite.T(). .Error(string(err.(*exec.ExitError).Stderr))
// 	}

// 	_, err = execOnMongo([]string{"rs.initiate({id: 'rs0', members: [ { _id: 0, host: 'localhost:27017' } ]})"})
// 	if err != nil {
// 		suite.T().Error(string(err.(*exec.ExitError).Stderr))
// 	}
// }

// func (suite *TestSuite) TearDownSuite() {
// 	_, err := exec.Command("/usr/bin/podman", []string{"stop", "mongodb"}...).Output()
// 	if err != nil {
// 		suite.T().Error(string(err.(*exec.ExitError).Stderr))
// 	}
// }

// func TestMongoUtilsHelperImpl_RunParallelOnShards_AllSuccess(t *testing.T) {
// 	command := "foobar"
// 	podNamePattern := "anypod%d"

// 	kubeHelperMock := mocks.NewKubernetesHelper(t)

// 	shards := 1

// 	for i := 0; i < shards; i++ {
// 		podname := fmt.Sprintf(podNamePattern, i)
// 		// ListPods(namespace string, labelSelectors map[string]string) (*corev1.PodList, error)
// 		kubeHelperMock.On("ListPods", mock.Anything, map[string]string{Microservice: fmt.Sprintf("datars%d", i+1)}).Return(&v1.PodList{Items: []v1.Pod{
// 			v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: podname, Namespace: ""}, Spec: v1.PodSpec{Containers: []v1.Container{v1.Container{Name: ""}}}}}}, nil)

// 		//log *zap.Logger, kubeConfig *rest.Config, podName string, namespace string, containerName string, command string, args []string) (string, error)
// 		kubeHelperMock.On("ExecRemote", mock.Anything, mock.Anything,
// 			podname, mock.Anything, mock.Anything, "bash", []string{" --eval \"rs.status().members.find((v)=>v.state==1).name.split('.')[0]\""}).Return(fmt.Sprintf("%s\n", podname), nil)

// 		kubeHelperMock.On("ExecRemote", mock.Anything, mock.Anything,
// 			podname, mock.Anything, mock.Anything, mock.Anything, []string{fmt.Sprintf(" --eval \"%s\"", command)}).Return("", nil)
// 	}

// 	r := &MongoUtilsHelperImpl{KubernetesHelperImpl: kubeHelperMock}
// 	err := r.RunParallelOnShards(command, shards)
// 	assert.NoError(t, err)
// }

// func (suite *TestSuite) TestMongoUtilsHelperImpl_RunParallelOnShards_AllError() {
// 	command := "foobar"
// 	podNamePattern := "anypod%d"

// 	kubeHelperMock := mocks.NewKubernetesHelper(suite.T())

// 	shards := 1

// 	for i := 0; i < shards; i++ {
// 		podname := fmt.Sprintf(podNamePattern, i)
// 		// ListPods(namespace string, labelSelectors map[string]string) (*corev1.PodList, error)
// 		kubeHelperMock.On("ListPods", mock.Anything, map[string]string{Microservice: fmt.Sprintf("datars%d", i+1)}).Return(&v1.PodList{Items: []v1.Pod{
// 			v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: podname, Namespace: ""}, Spec: v1.PodSpec{Containers: []v1.Container{v1.Container{Name: ""}}}}}}, nil)

// 		//log *zap.Logger, kubeConfig *rest.Config, podName string, namespace string, containerName string, command string, args []string) (string, error)
// 		kubeHelperMock.On("ExecRemote", mock.Anything, mock.Anything,
// 			podname, mock.Anything, mock.Anything, "bash", []string{" --eval \"rs.status().members.find((v)=>v.state==1).name.split('.')[0]\""}).Return(fmt.Sprintf("%s\n", podname), nil)

// 		kubeHelperMock.On("ExecRemote", mock.Anything, mock.Anything,
// 			podname, mock.Anything, mock.Anything, mock.Anything, []string{fmt.Sprintf(" --eval \"%s\"", command)}).Return("", fmt.Errorf("Some random error"))
// 	}

// 	r := &MongoUtilsHelperImpl{KubernetesHelperImpl: kubeHelperMock}
// 	err := r.RunParallelOnShards(command, shards)
// 	assert.Error(suite.T(), err)
// }

// // func (suite *TestSuite) TestMongoUtilsHelperImpl_CheckUserLogin() {
// // 	kubeHelperMock := mocks.NewKubernetesHelper(suite.T())

// // 	// ListPods(namespace string, labelSelectors map[string]string) (*corev1.PodList, error)
// // 	kubeHelperMock.On("ListPods", mock.Anything, mock.Anything).Return(&v1.PodList{Items: []v1.Pod{
// // 		v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "any", Namespace: ""}, Spec: v1.PodSpec{Containers: []v1.Container{v1.Container{Name: ""}}}}}}, nil)

// // 	kubeHelperMock.On("ExecRemote", mock.Anything, mock.Anything,
// // 		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
// // 		Return(
// // 			func(log *zap.Logger, kubeConfig *rest.Config, podName string, namespace string, containerName string, command string, args []string) string {
// // 				return ""
// // 			},
// // 			func(log *zap.Logger, kubeConfig *rest.Config, podName string, namespace string, containerName string, command string, args []string) error {
// // 				_, err := execOnMongo(args)
// // 				return err
// // 			})

// // 	r := &MongoUtilsHelperImpl{KubernetesHelperImpl: kubeHelperMock}
// // 	canLogin := r.CheckUserLogin("admin", "root", "root")
// // 	assert.False(suite.T(), canLogin)
// // }

// func execOnMongo(testCommand []string) ([]byte, error) {
// 	return exec.Command("bash", "-c", fmt.Sprintf("mongo %s", testCommand[0])).Output()
// }

// func (suite *TestSuite) TestMongoUtilsHelperImpl_RunOnMongoPod() {
// 	kubeHelperMock := mocks.NewKubernetesHelper(suite.T())

// 	kubeHelperMock.On("ExecRemote", mock.Anything, mock.Anything,
// 		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
// 		Return(
// 			func(log *zap.Logger, kubeConfig *rest.Config, podName string, namespace string, containerName string, command string, args []string) string {
// 				out, _ := execOnMongo(args)
// 				return string(out)
// 			},
// 			func(log *zap.Logger, kubeConfig *rest.Config, podName string, namespace string, containerName string, command string, args []string) error {
// 				// _, err := execOnMongo(args)
// 				return nil
// 			})

// 	r := &MongoUtilsHelperImpl{KubernetesHelperImpl: kubeHelperMock}
// 	r.Cmd = "admin --host localhost --quiet"
// 	out, err := r.RunOnMongoPod(getEmptyPod(), "db.version()")

// 	assert.NoError(suite.T(), err)
// 	assert.Equal(suite.T(), "4.2.23\n", out)
// }

// func getEmptyPod() *v1.Pod {
// 	return &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "any", Namespace: ""}, Spec: v1.PodSpec{Containers: []v1.Container{v1.Container{Name: ""}}}}
// }

// func (suite *TestSuite) TestMongoUtilsHelperImpl_CreateUser() {
// 	kubeHelperMock := mocks.NewKubernetesHelper(suite.T())

// 	podname := "foobar"

// 	// ListPods(namespace string, labelSelectors map[string]string) (*corev1.PodList, error)
// 	kubeHelperMock.On("ListPods", mock.Anything, map[string]string{Microservice: fmt.Sprintf("datars%d", 1)}).Return(&v1.PodList{Items: []v1.Pod{
// 		v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: podname, Namespace: ""}, Spec: v1.PodSpec{Containers: []v1.Container{v1.Container{Name: ""}}}}}}, nil)

// 	//log *zap.Logger, kubeConfig *rest.Config, podName string, namespace string, containerName string, command string, args []string) (string, error)
// 	kubeHelperMock.On("ExecRemote", mock.Anything, mock.Anything,
// 		podname, mock.Anything, mock.Anything, "bash", []string{" --eval \"rs.status().members.find((v)=>v.state==1).name.split('.')[0]\""}).Return(fmt.Sprintf("%s\n", podname), nil)

// 	kubeHelperMock.On("ExecRemote", mock.Anything, mock.Anything,
// 		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
// 		Return(
// 			func(log *zap.Logger, kubeConfig *rest.Config, podName string, namespace string, containerName string, command string, args []string) string {
// 				out, _ := execOnMongo(args)
// 				return string(out)
// 			},
// 			func(log *zap.Logger, kubeConfig *rest.Config, podName string, namespace string, containerName string, command string, args []string) error {
// 				// _, err := execOnMongo(args)
// 				return nil
// 			})

// 	r := &MongoUtilsHelperImpl{KubernetesHelperImpl: kubeHelperMock}
// 	r.Cmd = "admin --host localhost --quiet"
// 	err := r.CreateUser("admin", "root", "root", "'root'", false, false, true, 1)

// 	assert.NoError(suite.T(), err)
// }
