/*
Copyright 2022.

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

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	impl "github.com/Netcracker/qubership-mongodb-operator/pkg"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/mongodb"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/types"
)

// MongodbDeploymentReconciler reconciles a MongoService object
type MongodbDeploymentReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Reconciler reconcile.Reconciler
}

func (r *MongodbDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reconc, err := r.Reconciler.Reconcile(ctx, req)
	return reconc, err
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
func (r *MongodbDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Reconciler = newReconciler(mgr)
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.MongodbDeployment{}).
		WithEventFilter(ignoreStatusUpdatePredicate()).
		Complete(r)
}

func newReconciler(mgr ctrl.Manager) reconcile.Reconciler {
	return &core.ReconcileCommonService{
		Client:           mgr.GetClient(),
		KubeConfig:       mgr.GetConfig(),
		Scheme:           mgr.GetScheme(),
		Executor:         core.DefaultExecutor(),
		Builder:          &impl.MongoServiceBuilder{},
		PredeployBuilder: &impl.PreDeployBuilder{},
		DRBuilder:        &impl.DRBuilder{},
		Reconciler:       NewCommonReconciler(),
	}
}

// blank assignment to verify that ReconcileMongoService implements reconcile.Reconciler
var _ reconcile.Reconciler = &core.ReconcileCommonService{}

type MongodbDeploymentInstanceReconciler struct {
	Instance *v1alpha1.MongodbDeployment
}

func (s *MongodbDeploymentInstanceReconciler) GetConsulRegistration() *types.ConsulRegistration {
	return nil
}

func (s *MongodbDeploymentInstanceReconciler) GetConsulServiceRegistrations() map[string]*types.AgentServiceRegistration {
	return nil
}

func NewCommonReconciler() core.CommonReconciler {
	return &MongodbDeploymentInstanceReconciler{}
}

func (s *MongodbDeploymentInstanceReconciler) SetServiceInstance(client client.Client, request reconcile.Request) {
	mongoServiceList := &v1alpha1.MongodbDeploymentList{}
	err := core.ListRuntimeObjectsByNamespace(mongoServiceList, client, request.Namespace)
	if err != nil {
		if errors.IsNotFound(err) {

		}
	}
	msCount := len(mongoServiceList.Items)
	if msCount != 1 {
		//r.reqLogger.Error("There are " + fmt.Sprintf("%v", msCount) + " instances of Mongdobservice. Please leave only one.")
	}
	s.Instance = &mongoServiceList.Items[0]
}

func (s *MongodbDeploymentInstanceReconciler) UpdateStatus(condition types.ServiceStatusCondition) {
	s.Instance.Status.Conditions = []types.ServiceStatusCondition{condition}
}

func (s *MongodbDeploymentInstanceReconciler) UpdateDRStatus(drStatus types.DisasterRecoveryStatus) {
	drStatus.Mode = s.Instance.Spec.DisasterRecovery.Mode
	s.Instance.Status.DisasterRecoveryStatus = drStatus
}

func (s *MongodbDeploymentInstanceReconciler) GetStatus() *types.ServiceStatusCondition {
	if len(s.Instance.Status.Conditions) > 0 {
		return &s.Instance.Status.Conditions[0]
	}
	return nil
}

func (s *MongodbDeploymentInstanceReconciler) GetSpec() interface{} {
	return s.Instance.Spec
}

func (s *MongodbDeploymentInstanceReconciler) GetInstance() client.Object {
	return s.Instance
}

func (s *MongodbDeploymentInstanceReconciler) GetDeploymentVersion() string {
	return s.Instance.Spec.DeploymentVersion
}

func (s *MongodbDeploymentInstanceReconciler) GetVaultRegistration() *types.VaultRegistration {
	return &s.Instance.Spec.VaultRegistration
}

func (s *MongodbDeploymentInstanceReconciler) GetMessage() string {
	if len(s.Instance.Status.Conditions) > 0 {
		return s.Instance.Status.Conditions[0].Message
	}

	return ""
}

func (s *MongodbDeploymentInstanceReconciler) GetConfigMapName() string {
	return "mongodb-last-applied-configuration-info"
}

func (s *MongodbDeploymentInstanceReconciler) UpdatePassword() core.Executable {

	compound := mongodb.UpdateRootPasswordCompound{}
	compound.AddStep(&mongodb.UpdateMongoDBCredentials{})
	compound.AddStep(&mongodb.UpdateContextAuthMongo{})

	return &compound
}

func (s *MongodbDeploymentInstanceReconciler) GetAdminSecretName() string {
	return s.Instance.Spec.MongoRootSecretName
}

func (s *MongodbDeploymentInstanceReconciler) UpdatePassWithFullReconcile() bool {
	return false
}
