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
