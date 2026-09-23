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

package robotTests

import (
	"github.com/Netcracker/qubership-mongodb-supplementary/pkg/utils"
	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
)

// todo to core
type RobotCompound struct {
	core.MicroServiceCompound
}

type RobotBuilder struct {
	core.ExecutableBuilder
}

func (r *RobotBuilder) Build(ctx core.ExecutionContext) core.Executable {

	robot := RobotCompound{}
	robot.ServiceName = utils.Robot
	robot.CalcDeployType = func(ctx core.ExecutionContext) (deployType core.MicroServiceDeployType, err error) {
		return core.CleanDeploy, nil
	}
	robot.AddStep(&RobotDeployment{})

	return &robot
}
