package mongodb

import (
	"fmt"

	"github.com/Netcracker/qubership-mongodb-operator/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	coretypes "github.com/Netcracker/qubership-nosqldb-operator-core/pkg/types"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// WaitAndHandlePVCExpansionStep waits for in-progress PVC expansions to settle and restarts
// MongoDB pods when the storage class requires a filesystem resize (FileSystemResizePending).
type WaitAndHandlePVCExpansionStep struct {
	core.DefaultExecutable
	WaitTimeout int
}

func (r *WaitAndHandlePVCExpansionStep) Execute(ctx core.ExecutionContext) error {
	resizeNeeded, _ := ctx.Get(constants.PVCResizeNeeded).(bool)
	if !resizeNeeded {
		return nil
	}

	helperImpl := ctx.Get(utils.KubernetesHelperImpl).(core.KubernetesHelper)
	pvcNames := ctx.Get(utils.PvcNames).([]string)
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbDeployment)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)

	storage := spec.Spec.MongoDB.Storage

	anyNeedsRestart := false
	for i, pvcName := range pvcNames {
		desiredSize, err := desiredSizeForPVC(storage, i)
		if err != nil {
			return fmt.Errorf("failed to determine desired size for PVC %s: %w", pvcName, err)
		}

		log.Info(fmt.Sprintf("Waiting for PVC %s expansion to desired size %s", pvcName, desiredSize.String()))
		needsRestart, err := helperImpl.WaitForPVCExpansion(pvcName, request.Namespace, desiredSize, r.WaitTimeout)
		if err != nil {
			return fmt.Errorf("waiting for PVC %s expansion: %w", pvcName, err)
		}
		if needsRestart {
			anyNeedsRestart = true
		}
	}

	if !anyNeedsRestart {
		return nil
	}

	// FileSystemResizePending: filesystem resize completes only after pod restart.
	// Restart MongoDB pods one at a time so the replica set stays available.
	log.Info("PVC filesystem resize pending, performing rolling pod restart")
	pods, err := helperImpl.ListPods(request.Namespace, map[string]string{utils.Microservice: utils.MongoCluster})
	if err != nil {
		return fmt.Errorf("listing MongoDB pods for restart: %w", err)
	}
	for i := range pods.Items {
		pod := &pods.Items[i]
		log.Info(fmt.Sprintf("Restarting pod %s for filesystem resize completion", pod.Name))
		if err := helperImpl.RestartPod(pod, request.Namespace, r.WaitTimeout); err != nil {
			return fmt.Errorf("failed to restart pod %s: %w", pod.Name, err)
		}
	}
	return nil
}

func (r *WaitAndHandlePVCExpansionStep) Condition(ctx core.ExecutionContext) (bool, error) {
	return true, nil
}

// desiredSizeForPVC returns the storage size for PVC at position idx using the same
// modulo distribution as PVCTemplate in the core library.
func desiredSizeForPVC(storage *coretypes.StorageRequirements, idx int) (resource.Quantity, error) {
	if storage == nil || len(storage.Size) == 0 {
		return resource.Quantity{}, fmt.Errorf("storage size not configured")
	}
	sizeStr := storage.Size[idx%len(storage.Size)]
	qty, err := resource.ParseQuantity(sizeStr)
	if err != nil {
		return resource.Quantity{}, fmt.Errorf("invalid storage size %q: %w", sizeStr, err)
	}
	return qty, nil
}
