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

package pkg

import (
	"net/http"

	"github.com/Netcracker/qubership-dbaas-adapter-core/pkg/dao"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
)

// DescriptorPath is the declarative registration endpoint polled by the DBaaS Operator.
const DescriptorPath = "/api/v2/adapter/physical_database"

// PhysicalDatabaseDescriptor is the response of the declarative registration endpoint.
type PhysicalDatabaseDescriptor struct {
	PhysicalDatabaseId string            `json:"physicalDatabaseId"`
	Type               string            `json:"type"`
	Labels             map[string]string `json:"labels"`
	ApiVersions        dao.ApiVersions   `json:"apiVersions"`
	Features           map[string]bool   `json:"features"`
	SupportedRoles     []string          `json:"supportedRoles"`
	ReadOnlyHost       string            `json:"readOnlyHost,omitempty"`
}

// DescriptorSource supplies the adapter data that is only known at runtime.
type DescriptorSource interface {
	GetSupportedRoles() []string
	GetFeatures() map[string]bool
	GetROHost() string
}

// BuildDescriptor assembles the descriptor from the adapter's real configuration.
func BuildDescriptor(physicalDatabaseId, dbType string, labels map[string]string, src DescriptorSource) PhysicalDatabaseDescriptor {
	if labels == nil {
		labels = map[string]string{}
	}
	return PhysicalDatabaseDescriptor{
		PhysicalDatabaseId: physicalDatabaseId,
		Type:               dbType,
		Labels:             labels,
		ApiVersions: dao.ApiVersions{Specs: []dao.ApiVersionsSpec{{
			SpecRootUrl:     dao.RootUrl,
			Major:           dao.MajorAPIVersion,
			Minor:           dao.MinorAPIVersion,
			SupportedMajors: dao.SupportedMajorsVersions,
		}}},
		Features:       src.GetFeatures(),
		SupportedRoles: src.GetSupportedRoles(),
		ReadOnlyHost:   src.GetROHost(),
	}
}

// RegisterDescriptorHandler exposes the descriptor behind the adapter's existing Basic Auth credentials.
// The handler only reports adapter data; it never talks to the Aggregator.
func RegisterDescriptorHandler(app *fiber.App, user, pass string, descriptor PhysicalDatabaseDescriptor) {
	auth := basicauth.New(basicauth.Config{
		Realm: "This API is for using by dbaas operator only",
		Users: map[string]string{user: pass},
	})
	app.Get(DescriptorPath, auth, func(c *fiber.Ctx) error {
		if descriptor.PhysicalDatabaseId == "" || descriptor.Type == "" {
			return c.Status(http.StatusServiceUnavailable).SendString("Physical database identifier is not specified")
		}
		return c.JSON(descriptor)
	})
}
