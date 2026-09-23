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
	"fmt"

	"github.com/Netcracker/qubership-mongodb-supplementary/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type BackupConfigMaps struct {
	core.DefaultExecutable
}

func (r *BackupConfigMaps) Execute(ctx core.ExecutionContext) error {
	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
	log := ctx.Get(constants.ContextLogger).(*zap.Logger)

	log.Info("Backup Config Maps creation step started")

	cfg := &v1.ConfigMap{
		ObjectMeta: v12.ObjectMeta{
			Namespace: request.Namespace,
			Name:      utils.BackupMonitoringConfig,
		},
		Data: map[string]string{
			"url.health": fmt.Sprintf("http://%s:8080/health", ctx.Get(utils.MonitoringIPTemplate).(string)),
		},
	}

	err := utils.CreateRuntimeObjectContextWrapper(ctx, cfg, cfg.ObjectMeta)

	if err != nil {
		return err
	}

	log.Info("Backup Monitoring Config has been created")

	return nil
}
