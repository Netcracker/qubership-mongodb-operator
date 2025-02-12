package common

import (
	mFiber "github.com/Netcracker/qubership-mongodb-operator/pkg/fiber"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"go.uber.org/zap"
)

type RunFiberServer struct {
	core.DefaultExecutable
}

func (r *RunFiberServer) Execute(ctx core.ExecutionContext) error {
	log := ctx.Get(constants.ContextLogger).(*zap.Logger).Named("RunFiberServer")
	service := mFiber.NewMongoFiberService(ctx)
	service.UpdateCtx(ctx)

	return mFiber.RunFiberServer(8069, service, log)
}
