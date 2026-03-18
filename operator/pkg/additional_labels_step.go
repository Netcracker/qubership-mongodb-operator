package impl

import (
	//"context"
	//v1 "git.netcracker.com/PROD.Platform.Databases/mongodb-operator/api/v2"
	//"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	//"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	//"k8s.io/apimachinery/pkg/types"
	//"sigs.k8s.io/controller-runtime/pkg/client"
	//v1core "k8s.io/api/core/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
)

type SetAdditionalLabelsForProjectStep struct {
	core.DefaultExecutable
}

func (r *SetAdditionalLabelsForProjectStep) Execute(ctx core.ExecutionContext) error {
	//TODO: Add namespace patching
	//client := ctx.Get(constants.ContextClient).(client.Client)
	//request := ctx.Get(constants.ContextRequest).(reconcile.Request)

	//client.Update(context.TODO(), types.?????? ,&v1core.Namespace{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name: request.Namespace,
	//	},
	//})

	return nil
}
