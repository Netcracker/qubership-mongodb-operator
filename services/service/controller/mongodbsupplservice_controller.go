/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8type "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/Netcracker/qubership-mongodb-supplementary/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-supplementary/pkg"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/types"
)

var setupLog = ctrl.Log.WithName("setup")

// MongodbSupplServiceReconciler reconciles a MongodbSupplService object
type MongodbSupplServiceReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Reconciler reconcile.Reconciler
}

//+kubebuilder:rbac:groups=netcracker.com,resources=mongodbsupplservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=netcracker.com,resources=mongodbsupplservices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=netcracker.com,resources=mongodbsupplservices/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MongodbSupplService object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *MongodbSupplServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if delaySec := getEnvAsInt("MONGODB_OPERATOR_WAIT_DELAY_SECONDS", 30); delaySec > 0 {
		setupLog.Info("Applying initial delay before checking MongoDB Operator", "seconds", delaySec)
		time.Sleep(time.Duration(delaySec) * time.Second)
	}

	err := WaitForMongoDBOperatorReady(r.Client, "mongodb-operator", req.Namespace)
	if err != nil {
		setupLog.Info("MongoDB Operator not ready...")
		return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
	}

	reconc, err := r.Reconciler.Reconcile(ctx, req)
	return reconc, err
}

func getEnvAsInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func WaitForMongoDBOperatorReady(k8sClient client.Client, name, namespace string) error {
	setupLog.Info("Waiting for MongoDB CR to be ready...")
	ctx := context.Background()

	mongoGVK := schema.GroupVersionKind{
		Group:   "netcracker.com",
		Version: "v1alpha1",
		Kind:    "MongodbDeployment",
	}

	return wait.PollUntilContextTimeout(ctx, 5*time.Second, 10*time.Minute, true, func(ctx context.Context) (bool, error) {
		mongoCR := &unstructured.Unstructured{}
		mongoCR.SetGroupVersionKind(mongoGVK)

		err := k8sClient.Get(ctx, k8type.NamespacedName{
			Name:      name,
			Namespace: namespace,
		}, mongoCR)
		if err != nil {
			return false, err
		}

		status, found, err := unstructured.NestedSlice(mongoCR.Object, "status", "conditions")
		if !found || err != nil {
			return false, fmt.Errorf("unable to find status.conditions in MongoDBCluster")
		}

		for _, cond := range status {
			condMap, ok := cond.(map[string]interface{})
			if !ok {
				continue
			}

			t, _ := condMap["type"].(string)
			s, ok := condMap["status"].(bool) // read as bool instead of string

			if !ok {
				continue
			}

			switch strings.ToLower(t) {
			case "successful":
				if s {
					setupLog.Info("MongoDB CR status is Successful")
					return true, nil
				}
			case "failed":
				if s {
					setupLog.Error(nil, "MongoDB CR failed")
					return true, fmt.Errorf("MongoDB CR failed")
				}
			}
			setupLog.Info("Waiting for MongoDB CR to be ready", "type", t, "status", s)
		}

		return false, nil
	})
}

func ignoreStatusUpdatePredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *MongodbSupplServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Reconciler = newReconciler(mgr)
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.MongodbSupplService{}).
		WithEventFilter(ignoreStatusUpdatePredicate()).
		Complete(r)
}

func newReconciler(mgr ctrl.Manager) reconcile.Reconciler {
	return &core.ReconcileCommonService{
		Client:     mgr.GetClient(),
		KubeConfig: mgr.GetConfig(),
		Scheme:     mgr.GetScheme(),
		Executor:   core.DefaultExecutor(),
		Builder:    &pkg.MongoServiceBuilder{},
		// PredeployBuilder: &impl.PreDeployBuilder{},
		// DRBuilder:        &impl.DRBuilder{},
		Reconciler: NewCommonReconciler(),
	}
}

// blank assignment to verify that ReconcileMongoService implements reconcile.Reconciler
var _ reconcile.Reconciler = &core.ReconcileCommonService{}

type MongodbServiceInstanceReconciler struct {
	Instance *v1alpha1.MongodbSupplService
}

func (s *MongodbServiceInstanceReconciler) GetConsulRegistration() *types.ConsulRegistration {
	return nil
}

func (s *MongodbServiceInstanceReconciler) GetConsulServiceRegistrations() map[string]*types.AgentServiceRegistration {
	return nil
}

func NewCommonReconciler() core.CommonReconciler {
	return &MongodbServiceInstanceReconciler{}
}

func (s *MongodbServiceInstanceReconciler) SetServiceInstance(client client.Client, request reconcile.Request) {
	mongoServiceList := &v1alpha1.MongodbSupplServiceList{}
	err := core.ListRuntimeObjectsByNamespace(mongoServiceList, client, request.Namespace)
	if err != nil {
		if errors.IsNotFound(err) {

		}
	}
	msCount := len(mongoServiceList.Items)
	if msCount != 1 {
	}
	s.Instance = &mongoServiceList.Items[0]
}

func (s *MongodbServiceInstanceReconciler) UpdateStatus(condition types.ServiceStatusCondition) {
	s.Instance.Status.Conditions = []types.ServiceStatusCondition{condition}
}

func (s *MongodbServiceInstanceReconciler) GetConfigMapName() string {
	return "mongodb-services-last-applied-configuration-info"
}

func (s *MongodbServiceInstanceReconciler) UpdateDRStatus(drStatus types.DisasterRecoveryStatus) {
}

func (s *MongodbServiceInstanceReconciler) GetStatus() *types.ServiceStatusCondition {
	if len(s.Instance.Status.Conditions) > 0 {
		return &s.Instance.Status.Conditions[0]
	}
	return nil
}

func (s *MongodbServiceInstanceReconciler) GetSpec() interface{} {
	return s.Instance.Spec
}

func (s *MongodbServiceInstanceReconciler) GetInstance() client.Object {
	return s.Instance
}

func (s *MongodbServiceInstanceReconciler) GetDeploymentVersion() string {
	return s.Instance.Spec.DeploymentVersion
}

func (s *MongodbServiceInstanceReconciler) GetMessage() string {
	if len(s.Instance.Status.Conditions) > 0 {
		return s.Instance.Status.Conditions[0].Message
	}

	return ""
}

func (s *MongodbServiceInstanceReconciler) UpdatePassword() core.Executable {
	return nil
}

func (s *MongodbServiceInstanceReconciler) UpdatePassWithFullReconcile() bool {
	return true
}

func (s *MongodbServiceInstanceReconciler) GetAdminSecretName() string {
	return s.Instance.Spec.MongoDB.MongoRootSecretName
}
