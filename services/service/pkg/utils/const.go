package utils

const ContextMongoHost = "ContextMongoHost"
const ContextMongoService = "ContextMongoService"

const MongoDBDeploymentType = "MongoDBDeploymentType"
const MaxReplicaSize = "maxOf_CNF_DATA_replicaSize"
const MaxPVCCountForService = "max_pvc_count"
const PvcNames = "pvcNames"
const PVNodes = "pvNodeNames"
const BackupPvcNames = "pvcBackupNames"
const BackupPVNodes = "pvBackupNodeNames"
const ContextCredsManager = "contextCredsManager"

var BackupEntrypoint = []string{"python3", "/opt/backup/backup-daemon.py"}

const ArbiterIndexSelectorFunc = "contextArbSelectorFunc"

const MongoSecret = "mongo-secret"
const MongoSecretKeyFile = "mongodb-keyfile"
const MongoRootCreds = "mongodb-root-credentials.v1"

const RootCert = "root-ca"
const RootCertPath = "/usr/ssl/"
const ServerCertsPath = "/certs/"

const KubernetesHelperImpl = "kubernetesHelperImpl"

const MongoHelperImpl = "mongoHelperImpl"

const KubeHostName = "kubernetes.io/hostname"
const ServiceName = "serviceName"
const Name = "name"
const App = "app"
const Data = "data"
const Tmp = "tmp"
const MongoCluster = "mongo-cluster"
const Username = "username"
const Password = "password"
const Role = "role"
const AuthDatabase = "auth-database"

const ContextPasswordKey = "ctxPasswordKey"

const BashCommand = "bash"

const Microservice = "microservice"

const MongoPvcNameFormat = "pvc-%s-data"
const BackupPvcNameFormat = "mongodb-backup-storage"
const Backup = "backup"

const CnfNameKey = "cnfrs"
const CnfNameWithIndexFormat = CnfNameKey + "%v"

const DataNameKey = "datars%v"
const DataNameWithIndexesFormat = DataNameKey + "%v"

const MongoNode = "mongo-node"

const StatefulSetPodNameTemplate = "%s-0"
const MongoDomainTemplate = "%s.svc.cluster.local"
const MongoCMD = "%s %s --host=localhost --port=27017 --quiet"
const MongoCMDAuthTemplate = "%s mongodb://%s:%s@localhost:27017/%s --quiet"
const MongoCheckUserLogin = "%s mongodb://%s:%s@localhost:27017/%s --quiet --eval='quit()'"
const ContextMongoCMD = "contextMongoCmd"

const PrimaryReplicaCommand = "rs.status().members.find((v)=>v.state==1).name.split('.')[0]"
const ReplicaStateCommand = "rs.status().members.forEach(mem => print(mem.stateStr))"
const featureCompatibilityVersionCommand = "db.adminCommand({ getParameter: 1, featureCompatibilityVersion: 1 }).featureCompatibilityVersion.version"

const Mongos = "mongos"
const MongosPrivate = Mongos + "-private"

const RecyclerNameTemplate = "pv-recycler-pvc-%s"
const RecyclerPod = "recycler-pod"

const TriesCount = "triesCount"
const RetryTimeoutSec = "retryTimeout"

const BackupDaemon = "mongodb-backup-daemon"
const BackupStorage = "backup-storage"
const BackupConfigNodes = "backupConfigNodes"
const BackupBackupCreds = "mongodb-backup-credentials.v1"
const BackupRestoreCreds = "mongodb-restore-credentials.v1"
const BackupApiCreds = "mongodb-backup-api-credentials.v1"
const BackupMonitoringConfig = BackupDaemon + ".monitoring-config"

const ServicesUsersContextList = "ctxServicesUsersList"
const ServicesRolesContextList = "ctxServicesRolesList"

const Dbaas = "dbaas"
const DbaasName = "dbaas-mongo-adapter"
const DbaasMonitoringConfig = DbaasName + ".monitoring-config"
const DbaasAdminCreds = "mongo-dbaas-admin-credentials-secret.v1"
const DbaasAggregatorCreds = "dbaas-aggregator-credentials.v1"
const DbaasRegistrationCreds = "dbaas-aggregator-registration-credentials.v1"
const DbaasAdminRoleCreds = "dbaas-streaming-role.v1"

const MongoMonitoringAgent = "mongodb-monitoring-agent"
const MonitoringCreds = "mongodb-monitoring-credentials.v1"
const MonitoringContextVar = "mongodb-monitoring-password"
const MonitoringInfluxCreds = "mongodb-monitoring-influx-credentials"
const MongoMonitoringAgentSecret = MongoMonitoringAgent + "-secret"
const MongoMonitoringAgentSecretMountPath = "/etc/" + MongoMonitoringAgentSecret
const MonitroingViewRoleBinding = MongoMonitoringAgent + "-view-binding"
const MonitoringIPTemplate = "monitoringApiTemplate"

const MonitoringAgentAdditionalPrivilegesRoleName = "monitoringAgentAdditionalPrivileges"

const MongoPrometheusExporter = "mongodb-prometheus-exporter"
const PrometheusExporterCreds = "mongodb-prom-exporter-credentials.v1"

const SchemaSettingsValidationHAandArbiter = "Cnf Size & Data Size should be odd numbers. 3 or more"
const SchemaSettingsValidationDR = "Cnf Size & Data Size should be even numbers"
const SchemaSettingsValidationZeros = "Cnf Size, Data Size and Shard Count should be set"

const Policies = "policies"

const IsAnyCommonParameterChanged = "isAnyCommonParameterChanged"

const Robot = "robot-tests"
const RobotTestsFilesArgCheck = "ls ./output | wc -l"
const RobotTestsExpectedAnswer = "3"

const OperatorServiceName = "mongodb-services"

// user
const CreateORUpdateUserPattern = "" +
	"dbUsers = db.getSiblingDB('%[1]s').getUsers();" +
	"found = dbUsers.users?dbUsers.users.find((v)=>v.user=='%[2]s'):dbUsers.find((v)=>v.user=='%[2]s');" +
	"if (found) {" +
	"db.getSiblingDB('%[1]s').updateUser('%[2]s', {pwd: '%[3]s', roles: [%[4]s]})" +
	"} else {" +
	"db.getSiblingDB('%[1]s').createUser({user: '%[2]s', pwd: '%[3]s', roles: [%[4]s]});" +
	"}"

const CreateORUpdateRolePattern = "" +
	"dbRoles = db.getSiblingDB('%[1]s').getRoles();" +
	"found = dbRoles.roles?dbRoles.roles.find((v)=>v.role=='%[2]s'):dbRoles.find((v)=>v.role=='%[2]s');" +
	"if (found) {" +
	"db.getSiblingDB('%[1]s').updateRole('%[2]s', {privileges: [ %[3]s ], roles: [ %[4]s ]})" +
	"} else {" +
	"db.getSiblingDB('%[1]s').createRole({ role: '%[2]s', privileges: [ %[3]s ], roles: [ %[4]s ]});" +
	"}"

// dr
const ActiveMode = "active"
const StandbyMode = "standby"
const DisableMode = "disable"

const JsTemplate = "" +
	"var shouldBeHidden=function(member) { return [%s].indexOf(member.host) >= 0; };" +
	"var cfg = rs.config();" +
	"cfg.members = cfg.members.map(m => {" +
	"m.priority = shouldBeHidden(m) ? 0 : 1;" +
	"m.votes = shouldBeHidden(m) ? 0 : 1;" +
	"return m;" +
	"});" +
	"JSON.stringify(rs.reconfig(cfg, {force:true}, {writeConcern: {w:'majority'}}));"

const JsCountReplicasWithState = "rs.status().members.filter(mem => [%s].includes(mem.name)).reduce((prev, current) => {if (current.stateStr == '%s') return prev+1; else return prev;}, 0);"

const JsReplicasWithState = "rs.status().members.filter(mem => [%s].includes(mem.name)).map(x => x.stateStr)"

const JsReplicasWithStateAndName = "rs.status().members.map(function(m) { return {'name':m.name, 'stateStr':m.stateStr} })"

const shardHostsTemplate = "JSON.stringify(db.getSiblingDB('config').shards.updateOne( {'_id' : '%s'}, {\\$set:{ host: '%s/%s', state: 1}}))"

const shardTemplate = "" +
	"adminDB=db.getSiblingDB('local');" +
	"adminDB.dropDatabase();" +
	"adminDB=db.getSiblingDB('admin');" +
	"adminDB.system.version.updateOne({'_id' : 'shardIdentity', 'shardName' : '%s'}, { \\$set: {configsvrConnectionString: 'cnfrs/%s'}}, {writeConcern: { w:1, j:true }});" +
	"adminDB.system.version.remove({'_id': 'minOpTimeRecovery'}, {writeConcern: { w:1, j:true }})"

const versionCmd = "db.version()"

// http
const KeyFileURI = "keyfile"
const HealthURI = "healthz"
const RotateRolesURI = "rotate-roles"
const AddDRReplicasURI = "add-dr-replicas"
const FlushURI = "flush"
const CompactURI = "compact"
const RSStatusURI = "rs-status"

// status
const Up = "up"
const Down = "down"
const Degraded = "degraded"

const Charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

const MongoshBanner = `Warning: Could not access file: EACCES: permission denied, mkdir '/data/db/.mongodb'

Error: Could not open history file.
REPL session history will not be persisted.
 `

const (
	AppName              = "app.kubernetes.io/name"
	AppInstance          = "app.kubernetes.io/instance"
	AppVersion           = "app.kubernetes.io/version"
	AppComponent         = "app.kubernetes.io/component"
	AppManagedBy         = "app.kubernetes.io/managed-by"
	AppManagedByOperator = "app.kubernetes.io/managed-by-operator"
	AppProcByOperator    = "app.kubernetes.io/processed-by-operator"
	AppTechnology        = "app.kubernetes.io/technology"
	AppPartOf            = "app.kubernetes.io/part-of"
)
