package utils

import (
	"regexp"
	"strings"
)

const MongoDBDeploymentType = "MongoDBDeploymentType"
const MaxReplicaSize = "maxOf_CNF_DATA_replicaSize"
const MaxPVCCountForService = "max_pvc_count"
const PvcNames = "pvcNames"
const PVNodes = "pvNodeNames"
const BackupPvcNames = "pvcBackupNames"
const BackupPVNodes = "pvBackupNodeNames"

var BackupEntrypoint = []string{"python3", "/opt/backup/backup-daemon.py"}

const ArbiterIndexSelectorFunc = "contextArbSelectorFunc"

const MongoSecret = "mongo-secret"
const MongoSecretKeyFile = "mongodb-keyfile"
const MongoRootSecretName = "mongodb-root-credentials.v1"

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
const CloneModeType = "clone-mode-type"
const X509AuthMode = "x509"
const X509CnfrsCertificate = "cnfrs-x509-certificate"
const X509DatarsCertificate = "datars%d-x509-certificate"
const X509MongosCertificate = "mongos-x509-certificate"
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
const WriteConcernChangeCommand = "db.adminCommand({setDefaultRWConcern: 1,defaultWriteConcern: { w: 1 }})"
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
const BackupSecretName = "mongodb-backup-credentials.v1"
const RestoreUserSecretName = "mongodb-restore-credentials.v1"
const BackupApiSecretName = "mongodb-backup-api-credentials.v1"
const BackupMonitoringConfig = BackupDaemon + ".monitoring-config"

const ServicesUsersContextList = "ctxServicesUsersList"
const ServicesRolesContextList = "ctxServicesRolesList"

const Dbaas = "dbaas"
const DbaasName = "dbaas-mongo-adapter"
const DbaasMonitoringConfig = DbaasName + ".monitoring-config"
const DbaasAdminSecretName = "mongo-dbaas-admin-credentials-secret.v1"
const DbaasAggregatorSecretName = "dbaas-aggregator-credentials.v1"
const DbaasRegistrationSecretName = "dbaas-aggregator-registration-credentials.v1"
const DbaasAdminRoleSecretName = "dbaas-streaming-role.v1"

const DbaasPhysicalDatabasesLabels = "dbaas-physical-databases-labels"

const MongoMonitoringAgent = "mongodb-monitoring-agent"
const MonitoringSecretName = "mongodb-monitoring-credentials.v1"
const MonitoringContextVar = "mongodb-monitoring-password"
const MonitoringInfluxCreds = "mongodb-monitoring-influx-credentials"
const MongoMonitoringAgentSecret = MongoMonitoringAgent + "-secret"
const MongoMonitoringAgentSecretMountPath = "/etc/" + MongoMonitoringAgentSecret
const MonitroingViewRoleBinding = MongoMonitoringAgent + "-view-binding"
const MonitoringIPTemplate = "monitoringApiTemplate"

const MonitoringAgentAdditionalPrivilegesRoleName = "monitoringAgentAdditionalPrivileges"

const MongoPrometheusExporter = "mongodb-prometheus-exporter"
const PrometheusExporterSecretName = "mongodb-prom-exporter-credentials.v1"

const SchemaSettingsValidationHAandArbiter = "Cnf Size & Data Size should be odd numbers. 3 or more"
const SchemaSettingsValidationDR = "Cnf Size & Data Size should be even numbers"
const SchemaSettingsValidationZeros = "Cnf Size, Data Size and Shard Count should be set"

const Policies = "policies"

const IsAnyCommonParameterChanged = "isAnyCommonParameterChanged"

const Robot = "robot-tests"
const RobotTestsFilesArgCheck = "ls ./output | wc -l"
const RobotTestsExpectedAnswer = "3"

const OperatorServiceName = "mongodb-operator"

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
const UpdateRootPassword = "db.getSiblingDB('%s').updateUser('%s', {pwd: '%s'})"

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

const JsOpLogSize = "db.getSiblingDB('%s').oplog.rs.stats().maxSize"

const JsOpLogResize = "db.adminCommand({ replSetResizeOplog: 1, size: %v })"

const JsCheckMemberLag = `
(function(name) {
  var s = rs.status();
  if (!s.ok) return false;

  var p = s.members.find(m => m.stateStr === 'PRIMARY');
  var t = s.members.find(m => m.name === name);

  if (!p || !t) return false;

  var lag = (p.optimeDate - t.optimeDate) / 1000;

  return t.health === 1 &&
         t.stateStr === 'SECONDARY' &&
         t.optimeDate &&
         lag <= 30;
})('%s')`

const JsCheckMemberHealth = `
(function(name) {
  var s = rs.status();
  if (!s.ok) return false;

  var t = s.members.find(m => m.name === name);
  if (!t) return false;

  return t.health === 1 &&
         (t.stateStr === 'PRIMARY' || t.stateStr === 'SECONDARY');
})('%s')`

const JsCheckMemberExists = `
(function(host) {
  var cfg = rs.config();
  return cfg.members.some(function(m) { return m.host === host; });
})('%s')`

const JsCheckReconfigNeeded = `
(function(hiddenHosts) {
  var cfg = rs.config();
  var needed = false;
  cfg.members.forEach(function(m) {
    var shouldBeHidden = hiddenHosts.indexOf(m.host) >= 0;
    if (shouldBeHidden && (m.priority !== 0 || m.votes !== 0)) needed = true;
    if (!shouldBeHidden && (m.priority !== 1 || m.votes !== 1)) needed = true;
  });
  return needed;
})([%s])`

const JsCheckAllMembersLag = `
(function(names, maxLagSec) {
  var s = rs.status();
  if (!s || !s.ok) return false;
  var p = s.members.find(function(m) { return m.stateStr === 'PRIMARY'; });
  if (!p) return false;
  for (var i = 0; i < names.length; i++) {
    var t = s.members.find(function(m) { return m.name === names[i]; });
    if (!t || t.health !== 1 || t.stateStr !== 'SECONDARY' || !t.optimeDate) return false;
    if ((p.optimeDate - t.optimeDate) / 1000 > maxLagSec) return false;
  }
  return true;
})([%s], %d)`

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

const MongoshBanner7 = `Warning: Could not access file: EACCES: permission denied, mkdir '/data/db/.mongodb'

Error: Could not open history file.
REPL session history will not be persisted.
 `

const MongoshBanner8 = `Warning: Could not access file: EACCES: permission denied, mkdir '/data/db/.mongodb'`

func GetSetFeatureCompatibilityVersion(version string) string {
	if strings.HasPrefix(version, "7") || strings.HasPrefix(version, "8") {
		return "db.adminCommand( { setFeatureCompatibilityVersion: '%s' , confirm: true} )"
	} else {
		return "db.adminCommand( { setFeatureCompatibilityVersion: '%s' } )"
	}
}

// TODO drop when 4.4 dropped

func GetMongoImageVersion(dockerImage string) string {
	// Match a tag with optional underscore-separated suffix
	re := regexp.MustCompile(`^(?:.+/)?(?:mongo|.+mongodb):(?:.*_)?(\d+)\.\d+\.\d+$`)
	version := re.FindStringSubmatch(dockerImage)
	if len(version) < 2 {
		return "4" // default fallback
	}
	return version[1]
}

func MongoBinary(dockerImage string) string {
	return "mongosh"
}
