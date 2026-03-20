import json
import sys
import time
from pymongo import MongoClient, errors
from pymongo.database import Database
from robot.api import logger


def executed(operation_fun):
    """
    This function is used as decorator to execute mongoDB commands
    with opening and closing connections.
    :param operation_fun: the function which will be decorated
    :return: wrapped function
    """
    def wrapped(*args, **kargs):
        self = args[0]
        host_repl = kargs.get('host_repl')
        database_name = kargs.get('database_name', self._database_name)
        logger.debug(f"Executing {operation_fun.__name__} on DB={database_name}")
        self.connect_to_mongodb()
        self._database = self._client[database_name]
        if host_repl is not None:
            self.connect_to_mongodb_replicas(host_repl)
        try:
            return operation_fun(*args, **kargs)
        except Exception as e:
            logger.error(f"Exception occurred when {operation_fun.__name__} was executed: {str(e)}")
            raise Exception(e)
        finally:
            try:
                self.disconnect_from_mongodb()
            except Exception as e:
                logger.warn(f"Disconnect error after {operation_fun.__name__}: {e}")
    return wrapped


class MongoDBLibrary(object):
    """
    MongoDB testing library for Robot Framework.
    """

    ROBOT_LIBRARY_SCOPE = 'GLOBAL'

    def __init__(self, host, port, user, password, database_name, req_timeout_sec=20, host_datars="datars1", tls=False, tlsCAFile=""):
        self._host = host
        self._host_datars = host_datars
        self._port = int(port)
        self._user = user
        self._password = password
        self._database_name = database_name
        self._client = None  # type: MongoClient
        self._database = None  # type: Database
        self._timeoutMS = int(req_timeout_sec) / 5 * 1000
        self._tls = tls
        self._tlsCAFile = tlsCAFile

    def connect_to_mongodb(self):
        """
        Connects to MongoDB.
        *Example:*\n
            | Connect To Mongodb |
        """
        try:
            if self._tls:
                self._client = MongoClient(host=self._host, port=self._port, username=self._user, password=self._password, connectTimeoutMS=self._timeoutMS, socketTimeoutMS=self._timeoutMS, serverSelectionTimeoutMS=self._timeoutMS, tls=self._tls, tlsCAFile=self._tlsCAFile, tlsAllowInvalidCertificates=True)
            else:
                self._client = MongoClient(host=self._host, port=self._port, username=self._user, password=self._password, connectTimeoutMS=self._timeoutMS, socketTimeoutMS=self._timeoutMS, serverSelectionTimeoutMS=self._timeoutMS)
        except Exception as e:
            raise Exception(f'Connect to MongoDB  Error: {e}')

    def connect_to_mongodb_replicas(self, host_repl):
        """
        Connects to MongoDB Datars.
        *Example:*\n
            | Connect To Mongodb Datars|
        """
        try:
            if self._tls:
                self._host_repl = MongoClient(host=host_repl, port=self._port, username=self._user, password=self._password, connectTimeoutMS=self._timeoutMS, socketTimeoutMS=self._timeoutMS, serverSelectionTimeoutMS=self._timeoutMS, tls=self._tls, tlsCAFile=self._tlsCAFile, tlsAllowInvalidCertificates=True)
            else:
                self._host_repl = MongoClient(host=host_repl, port=self._port, username=self._user, password=self._password, connectTimeoutMS=self._timeoutMS, socketTimeoutMS=self._timeoutMS, serverSelectionTimeoutMS=self._timeoutMS)
        except Exception as e:
            raise Exception(f'Connect to MongoDB Replica Error: {e}')


    @executed
    def drop_mongodb_database(self, database_name, **kwargs):
        """
        Drops the database if it exists.
        *Args:*\n
            _database_name_ - database name;\n
        *Example:*\n
            | Drop Mongodb Database | robot_database |
        """
        self._client.drop_database(database_name)

    @executed
    def drop_mongodb_collection(self, collection_name, **kwargs):
        """
        Drops the collection if it exists.
        By default it uses current database_name
        but it can be specified.
        A connection will be created and closed.
        *Args:*\n
            collection_name - collection name;\n
        *Kwargs:*\n
            database_name - database name;\n
        *Example:*\n
            | Drop Mongodb Collection | robot_collection |
            | Drop Mongodb Collection | robot_collection | database_name=robot_database |
        """
        self._database.drop_collection(collection_name)

    @executed
    def insert_one_mongodb_document(self, collection_name, document, **kwargs):
        """
        Inserts the document into the collection.
        By default it uses current database_name
        but it can be specified.
        A connection will be created and closed.
        *Args:*\n
            collection_name - collection name;\n
            document - document;\n
        *Kwargs:*\n
            database_name - database name;\n
            collection_name - collection name;\n
        *Example:*\n
            | Insert One Mongodb Document | robot_collection | {"name" : "Tom"} |
            | Insert One Mongodb Document | robot_collection | {"name" : "Tom"} | database_name=robot_database |
        """
        collection = self._database[collection_name]
        collection.insert_one(json.loads(document))

    @executed
    def delete_one_mongodb_document(self, collection_name, _filter, **kwargs):
        """
        Delete the document from the collection.
        By default it uses current database_name
        but it can be specified.
        A connection will be created and closed.
        *Args:*\n
            collection_name - collection name;\n
            _filter - document field or fields which used to find document;\n
        *Kwargs:*\n
            database_name - database name;\n
        *Example:*\n
            | Delete One Mongodb Document | robot_collection | {"name" : "Tom"} |
            | Delete One Mongodb Document | robot_collection | {"name" : "Tom"} | database_name=robot_database |
        """
        collection = self._database[collection_name]
        collection.delete_one(json.loads(_filter))

    @executed
    def update_one_mongodb_document(self, collection_name, _filter, update, operator='set', **kwargs):
        """
        Update the document within the collection.
        By default it uses current database_name
        but it can be specified.
        A connection will be created and closed.
        *Args:*\n
            collection_name - collection name;\n
            _filter - a condition to choose a document within a collection;\n
            update - update json which will be used with $set predicate
            operator - MongoDB operator (set, inc, unset, etc.)
        *Kwargs:*\n
            database_name - database name;\n
        *Example:*\n
            | Update One Mongodb Document | robot_collection | {"name" : "Tom"} | {"key": "value"} | {"key": "value_updated"} |
            | Update One Mongodb Document | robot_collection | {"name" : "Tom"} | {"key": "value"} | {"key": "value_updated"} | database_name=robot_database |
        """
        collection = self._database[collection_name]
        collection.update_one(json.loads(_filter), json.loads(f'{{\"${operator}\":{update}}}'))

    @executed
    def check_document_exists(self, collection_name, document, **kwargs):
        result = self._database[collection_name].count_documents(json.loads(document), limit = 1) != 0
        logger.debug(f"Document exists: {result}")
        return result

    @executed
    def get_collection_names(self, **kwargs):
        result = self._database.list_collection_names()
        logger.debug(result)
        return result

    def disconnect_from_mongodb(self):
        """
        Closes all MongoDB clients.
        *Example:*\n
            | Connect To Mongodb |
            | Disconnect From Mongodb |
        """
        try:
            if self._client:
                self._client.close()
                logger.debug("Closed primary MongoDB client.")
            if hasattr(self, "_host_repl") and self._host_repl:
                self._host_repl.close()
                logger.debug("Closed replica MongoDB client.")
        except Exception as e:
            logger.warn(f"Error closing MongoDB connection: {e}")
        finally:
            self._client = None

    @executed
    def get_rs_status(self, **kwargs):
        """Get primary and secondary members, retry until election settles."""
        max_wait = 60
        start = time.time()
        while True:
            try:
                all_replicas = self._host_repl.admin.command('replSetGetStatus')
                replicas_secondary = []
                replicas_primary = []
                for repl in all_replicas.get('members', []):
                    state = repl.get('stateStr')
                    if state == 'SECONDARY':
                        replicas_secondary.append(repl['name'])
                    elif state == 'PRIMARY':
                        replicas_primary.append(repl['name'])

                if replicas_primary or (time.time() - start > max_wait):
                    logger.debug(f"RS Status => Primary: {replicas_primary}, Secondary: {replicas_secondary}")
                    return replicas_primary, replicas_secondary

                logger.info("Waiting for replica set election to complete...")
                time.sleep(5)
            except errors.AutoReconnect:
                logger.info("Replica set not ready yet, retrying...")
                time.sleep(5)
            except Exception as e:
                logger.warn(f"Error fetching replSetGetStatus: {e}")
                time.sleep(5)

    def check_reelection_primary(self, old_primary, new_primary):
        """Safely check if primary was re-elected."""
        try:
            old_name = old_primary[0] if old_primary else None
            new_name = new_primary[0] if new_primary else None
            if not new_name:
                logger.warn("No new primary detected yet.")
                return False
            if old_name != new_name:
                logger.debug(f"Primary changed from {old_name} to {new_name}")
                return True
            return False
        except Exception as e:
            logger.warn(f"Error checking reelection: {e}")
            return False

    @executed
    def check_replication(self, scheme, **kwargs):
        all_replicas = self._host_repl.admin.command('replSetGetStatus')
        members = all_replicas['members']
        state_success = sum(1 for r in members if r['state'] in (1, 2))
        count = len(members)
        if scheme == 'failover' and state_success >= count / 2:
            return True
        elif scheme == 'not_failover' and state_success == count:
            return True
        return False

    def check_location_component(self, main_os, left_pods, right_pods, mode='single'):
        if mode == 'single':
            if main_os == 'left' and left_pods > 0:
                return True
            elif main_os == 'right' and right_pods > 0:
                return True
            else:
                return False
        elif mode == 'plural_failover':
            if main_os == 'left' and left_pods > 2:
                return True
            elif main_os == 'right' and right_pods > 2:
                return True
            else:
                return False
        else:
            if main_os == 'left' and left_pods > 2 and right_pods > 2:
                return True
            elif main_os == 'right' and right_pods > 2 and left_pods > 2:
                return True
            else:
                return False

    def list_of_datars_hosts(self):
        hosts = self._host_datars.split(",")
        return hosts

    def get_multi_users_name(self, connectionProperties):
        users_roles = {}
        for con in connectionProperties:
            users_roles[con['role']] = con['username']
        return users_roles

    @executed
    def delete_db_user(self, db_name, user_name):
        db = self._client[db_name]
        db.command('dropUser', user_name)

    @executed
    def revoke_roles_from_user(self, db_name, user_name, role):
        db = self._client[db_name]
        db.command("revokeRolesFromUser", user_name, roles=[{"role": role, "db": db_name}])

    @executed
    def get_all_users_names_for_db(self, db_name):
        all_users = []
        db = self._client[db_name]
        users_info = db.command('usersInfo')
        for user in users_info['users']:
            all_users.append(user['user'])
        return all_users

    @executed
    def get_permission_for_role(self, db_name, role):
        permissions = []
        db = self._client[db_name]
        users_info = db.command('usersInfo')
        for user in users_info['users']:
            if user['user'] == role:
                for role in user['roles']:
                    permissions.append(role['role'])
                return permissions
        return None
