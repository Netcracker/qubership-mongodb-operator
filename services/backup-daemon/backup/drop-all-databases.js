// drop all mongo databases

// databases that crucial for mongo health
var systemDatabases = [
		"admin", // will be overwritten on import if it needed
		"config", // sharded cluster configuration
		"local" // ??
	];

db.getMongo().getDBNames().filter( dbName =>
		systemDatabases.indexOf(dbName) == -1
	).forEach( dbName => {
			print("Drop database: " + dbName);
			db.getMongo().getDB(dbName).dropDatabase();
		}
	);

var masterInfo = db.isMaster();
if(masterInfo['msg'] && masterInfo['msg'] === 'isdbgrid') {
    db.adminCommand("flushRouterConfig");
}