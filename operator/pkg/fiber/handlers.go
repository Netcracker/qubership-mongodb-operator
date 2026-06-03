package fiber

// To generate swagger spec run: swag init -g handlers.go

import (
	"fmt"
	"net/http"
	"runtime/debug"

	_ "github.com/Netcracker/qubership-mongodb-operator/pkg/fiber/docs"
	mUtils "github.com/Netcracker/qubership-mongodb-operator/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/gofiber/swagger"
	"go.uber.org/zap"
)

type MongoOperatorHandler struct {
	service MongoFiberService
	logger  *zap.Logger
}

// GetMongoKeyFile godoc
// @Tags Operator API
// @Summary Get Mongo Key File
// @Description Returns list of all existing databases in all namespaces
// @Produce  json
// @Success 200 {object} string
// @Failure 500 {string} Token "Internal error"
// @Router /keyfile [get]
func (h *MongoOperatorHandler) GetMongoKeyFile(c *fiber.Ctx) error {
	return c.JSON(h.service.GetKeyFile())
}

// Health godoc
// @Tags Operator API
// @Summary Get Mongo Cluster Health
// @Description Returns up/down/degraded health status
// @Produce  json
// @Success 200 {object} Status
// @Failure 500 {string} Token "Internal error"
// @Router /healthz [get]
func (h *MongoOperatorHandler) Health(c *fiber.Ctx) error {
	return c.JSON(h.service.Health())
}

// GetRSStatus godoc
// @Tags Operator API
// @Summary Get Mongo Replica sets status
// @Description Returns output of rs.status()
// @Produce  json
// @Success 200 {object} []map[string]interface{}
// @Failure 500 {string} Token "Internal error"
// @Router /rs-status [get]
func (h *MongoOperatorHandler) GetRSStatus(c *fiber.Ctx) error {
	return c.JSON(h.service.GetRSStatus())
}

func (h *MongoOperatorHandler) AddDRReplicas(c *fiber.Ctx) error {
	err := h.service.AddDRReplicas()
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}
	return c.SendStatus(http.StatusOK)
}

// FlushInMemoryData godoc
// @Tags Operator API
// @Summary Flush data
// @Description Flushes data from RAM to disks
// @Produce  json
// @Success 200 {object} []string
// @Failure 500 {string} Token "Internal error"
// @Router /flush [post]
func (h *MongoOperatorHandler) FlushInMemoryData(c *fiber.Ctx) error {
	err := c.JSON(h.service.FlushInMemoryData())
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}
	return c.SendStatus(http.StatusOK)
}

// Compact godoc
// @Tags Operator API
// @Summary Compact
// @Description Compacts the collection in the database
// @Produce  json
// @Success 200 {object} []string
// @Failure 500 {string} Token "Internal error"
// @Param dbName path string true "Databases to compact collection in"
// @Param collectionName path string true "Collection name"
// @Router /compact/{dbName}/{collectionName} [post]
func (h *MongoOperatorHandler) Compact(c *fiber.Ctx) error {
	dbName := c.Params("dbName")
	collectionName := c.Params("collectionName")
	err := c.JSON(h.service.Compact(dbName, collectionName))
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}
	return c.SendStatus(http.StatusOK)
}

// Compact godoc
// @Tags Operator API
// @Summary Compact all
// @Description Compacts all collections in the database
// @Produce  json
// @Success 200 {object} []string
// @Failure 500 {string} Token "Internal error"
// @Param dbName path string true "Databases to compact all collections in"
// @Router /compact/{dbName} [post]
func (h *MongoOperatorHandler) CompactAll(c *fiber.Ctx) error {
	dbName := c.Params("dbName")
	err := c.JSON(h.service.CompactAll(dbName))
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString(err.Error())
	}
	return c.SendStatus(http.StatusOK)
}

func SetMongoOperatorHandlers(app *fiber.App, service MongoFiberService, logger *zap.Logger) {
	handler := &MongoOperatorHandler{service: service, logger: logger}

	recoverConfig := recover.ConfigDefault
	recoverConfig.EnableStackTrace = true
	recoverConfig.StackTraceHandler = func(c *fiber.Ctx, e interface{}) {
		logger.Error(fmt.Sprintf("Panic: %+v\nStacktrace:\n%s", e, string(debug.Stack())))
	}
	app.Use(recover.New(recoverConfig))
	app.Use(func(c *fiber.Ctx) error {
		// Setting defaults for existed handlers
		c.Request().Header.SetContentType(utils.GetMIME("json"))
		logger.Debug(fmt.Sprintf("%s %s", c.Request().Header.Method(), c.Path()))
		return c.Next()
	})

	app.Get(fmt.Sprintf("/%s", mUtils.KeyFileURI), handler.GetMongoKeyFile)
	app.Get(fmt.Sprintf("/%s", mUtils.HealthURI), handler.Health)
	app.Get(fmt.Sprintf("/%s", mUtils.RSStatusURI), handler.GetRSStatus)
	app.Post(fmt.Sprintf("/%s", mUtils.AddDRReplicasURI), handler.AddDRReplicas)
	app.Post(fmt.Sprintf("/%s", mUtils.FlushURI), handler.FlushInMemoryData)
	app.Post(fmt.Sprintf("/%s/:dbName/:collectionName", mUtils.CompactURI), handler.Compact)
	app.Post(fmt.Sprintf("/%s/:dbName", mUtils.CompactURI), handler.CompactAll)

	app.Get("/swagger/*", swagger.New(swagger.Config{ // custom
		URL:         "/swagger/doc.json",
		DeepLinking: false,
	}))

}
