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

package fiber

import (
	"context"
	"sync"
	"time"

	nosqlFiber "github.com/Netcracker/qubership-nosqldb-operator-core/pkg/fiber"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

var runFiberServerOnce sync.Once

func RunFiberServer(serverPort int, service MongoFiberService, log *zap.Logger) error {
	var err error
	runFiberServerOnce.Do(func() {
		err = (*nosqlFiber.GetFiberService()).Create(serverPort, func(app *fiber.App, serverContext context.Context) error {
			app.Server().ReadTimeout = time.Duration(60 * time.Second)
			app.Server().WriteTimeout = time.Duration(60 * time.Second)
			app.Server().IdleTimeout = time.Duration(60 * time.Second)
			SetMongoOperatorHandlers(app, service, log)
			return nil
		}, true)
	})

	return err
}
