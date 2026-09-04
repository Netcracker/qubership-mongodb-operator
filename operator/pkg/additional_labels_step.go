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
