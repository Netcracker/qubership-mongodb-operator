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

package backup

import (
	"github.com/Netcracker/qubership-mongodb-supplementary/api/v1alpha1"
	"github.com/Netcracker/qubership-mongodb-supplementary/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
)

type PrepareContextForBackupService struct {
	core.DefaultExecutable
}

func (r *PrepareContextForBackupService) Execute(ctx core.ExecutionContext) error {
	spec := ctx.Get(constants.ContextSpec).(*v1alpha1.MongodbSupplService)

	ctx.Set(utils.MaxPVCCountForService, 1)

	if spec.Spec.SchemaSettings.SchemaType == v1alpha1.Single {
		ctx.Set(utils.BackupConfigNodes, 0)
	} else {
		ctx.Set(utils.BackupConfigNodes, spec.Spec.SchemaSettings.CnfReplicaSize)
	}

	return nil
}
