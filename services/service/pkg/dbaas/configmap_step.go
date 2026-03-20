package dbaas

// import (
// 	"encoding/json"
// 	"fmt"
// 	"strconv"
// 	"strings"

// 	"github.com/Netcracker/qubership-mongodb-supplementary/api/v1alpha1"
// 	"github.com/Netcracker/qubership-mongodb-supplementary/pkg/utils"
// 	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/constants"
// 	"github.com/Netcracker/qubership-nosqldb-operator-core/pkg/core"
// 	"go.uber.org/zap"
// 	v1 "k8s.io/api/core/v1"
// 	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"sigs.k8s.io/controller-runtime/pkg/reconcile"
// )

// type DbaasConfigMaps struct {
// 	core.DefaultExecutable
// }

// func (r *DbaasConfigMaps) Execute(ctx core.ExecutionContext) error {
// 	request := ctx.Get(constants.ContextRequest).(reconcile.Request)
// 	//mongoImpl := ctx.Get(utils.MongoHelperImpl).(utils.MongoHelper)
// 	log := ctx.Get(constants.ContextLogger).(*zap.Logger)
// 	spec := ctx.Get(constants.ContextSpec).(*v13.MongoService)
// 	dbaas := spec.Spec.Dbaas

// 	if spec.Spec.PrometheusExporter.Install {
// 		ipTemplate := ctx.Get(utils.MonitoringIPTemplate).(string)
// 		log.Info("Dbaas Config Maps creation step started")

// 		cfg := &v1.ConfigMap{
// 			ObjectMeta: v12.ObjectMeta{
// 				Namespace: request.Namespace,
// 				Name:      utils.DbaasMonitoringConfig,
// 			},
// 			Data: map[string]string{
// 				"url.health":                       fmt.Sprintf("http://%s:8080/health", ipTemplate),
// 				"prometheus.metrics":               fmt.Sprintf("http://%s:8080/prometheus", ipTemplate),
// 				"group.metrics.counter.status.all": "^counter.status..+",
// 				"group.metrics.gauge.response.all": "^gauge.response..+",
// 			},
// 		}

// 		err := utils.CreateRuntimeObjectContextWrapper(ctx, cfg, cfg.ObjectMeta)

// 		if err != nil {
// 			return err
// 		}

// 		log.Info("Dbaas Monitoring Config has been created")
// 	}

// 	log.Debug("Checking mongo version...")

// 	//  TODO  getMongoVersion

// 	// result, err := mongoImpl.GetMongoVersion()

// 	// if err != nil {
// 	// 	return &core.ExecutionError{Msg: "Can't get mongo version. Error: " + err.Error()}
// 	// }

// 	//TODO do we need this?

// 	result := "4.2.23"

// 	result = strings.TrimSpace(result)
// 	result = strings.TrimSuffix(result, "\n") // Workaround

// 	log.Debug("Mongo version is: " + result)

// 	dataMap := core.ConcatMaps(
// 		map[string]string{
// 			"MONGO_RELEASE": spec.Spec.MongoDB.DockerImage,
// 			"MONGO_VERSION": result,
// 			"IS_DR":         strconv.FormatBool(spec.Spec.SchemaSettings.SchemaType == v1alpha1.DR),
// 			"IS_HA":         strconv.FormatBool(spec.Spec.SchemaSettings.SchemaType != v1alpha1.Single),
// 			"NAMESPACE":     request.Namespace,
// 		},
// 		dbaas.DbaasPhysicalDatabasesCustomLabels)

// 	jsonString, jsonErr := json.Marshal(dataMap)

// 	if jsonErr != nil {
// 		return &core.ExecutionError{Msg: "Can't parse dbaas configuration. Error: " + jsonErr.Error()}
// 	}

// 	dataString := string(jsonString)

// 	cfg := &v1.ConfigMap{
// 		ObjectMeta: v12.ObjectMeta{
// 			Namespace: request.Namespace,
// 			Name:      utils.DbaasPhysicalDatabasesLabels,
// 		},
// 		Data: map[string]string{
// 			"dbaas.physical_databases.registration.labels.json": dataString,
// 		},
// 	}

// 	err := utils.CreateRuntimeObjectContextWrapper(ctx, cfg, cfg.ObjectMeta)

// 	if err != nil {
// 		return err
// 	}

// 	log.Info("Dbaas Physical Databases Labels Config has been created")

// 	return nil
// }
