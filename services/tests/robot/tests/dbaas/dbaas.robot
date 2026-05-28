*** Variables ***
${DBAAS_HOST}                                     %{DBAAS_HOST}
${ATTEMPTS_NUMBER}                                100

*** Settings ***
Library  String
Library	 Collections
Library	 RequestsLibrary
Library  OperatingSystem
Library  ../lib/Helper.py
Library  ../lib/KubernetesClient.py
Resource  ../shared/keywords.robot
Resource    dbaas-shared.robot
Suite Setup  Preparation
Suite Teardown  Cleanup

*** Keywords ***
Preparation
    Prepare Shared
    Preparation dbaas shared

    ${namePrefix}=  Catenate    SEPARATOR=   ${MONGO_DB}
    Set Suite Variable  ${namePrefix}

    ${resultDbName}=  Catenate    SEPARATOR=   ${MONGO_DB}
    Set Suite Variable  ${resultDbName}

    ${SECOND_MONGO_DB}=  Set Variable  granular_${MONGO_DB}
    Set Suite Variable  ${SECOND_MONGO_DB}

    ${collection}=  Generate Random String  10  [LOWER]
    Set Suite Variable  ${collection}

    ${docValueOne}=  Generate Random String  10  [LOWER]
    ${docValueTwo}=  Generate Random String  10  [LOWER]

    ${documentOne}=  Set Variable  {"name":"${docValueOne}"}
    ${documentTwo}=  Set Variable  {"name":"${docValueTwo}"}
    Set Suite Variable  ${documentOne}
    Set Suite Variable  ${documentTwo}

    ${metaCollection}=  Set Variable  _dbaas_metadata
    Set Suite Variable  ${metaCollection}

    ${meta_string}=  Generate Random String  10
    ${meta_value}=  Generate Random String  10

    Set Suite Variable  ${meta_string}
    Set Suite Variable  ${meta_value}

    ${agg_username} =    Encode URL  ${DBAAS_AGGREGATOR_USERNAME}
    ${agg_password} =    Encode URL  ${DBAAS_AGGREGATOR_PASSWORD}

    Prepare Configuration For Dbaas Connection
    Create Session    dbaassession    ${PROTOCOL}://${agg_username}:${agg_password}@${DBAAS_HOST}:${port}    verify=${verify}

Prepare Configuration For Dbaas Connection
    ${verify}=    Get Environment Variable    name=TLS_ROOTCERT    default=False
    ${dbaas_tls}=    Get Environment Variable    name=TLS_ENABLED    default=False
    ${port}=    Get Environment Variable    name=PORT    default=8080
    ${env_dbaas_aggregator_host}=  Create List  DBAAS_AGGREGATOR_REGISTRATION_ADDRESS
    ${dbaas_aggregator_host}=  Get Environment Variables For Deployment Entity Container  dbaas-mongo-adapter  ${NAMESPACE}  dbaas-mongo-adapter  ${env_dbaas_aggregator_host}
    ${https_aggregator_enabled}=  Evaluate  "https" in "${dbaas_aggregator_host}"
    Set Suite Variable  ${https_aggregator_enabled}
    @{connection_settings}=  Run Keyword If  '${https_aggregator_enabled}' == '${True}' and '${dbaas_tls}' == 'true'
    ...  Set Variable  https  ${port}  ${verify}
    ...  ELSE
    ...  Set Variable  http  8080  False
    log to console  PROTOCOL, PORT and VERIFY for DBAAS Connection: ${connection_settings[0]} ${connection_settings[1]} ${connection_settings[2]}
    Set Suite Variable  ${verify}  ${connection_settings[2]}
    Set Suite Variable  ${port}  ${connection_settings[1]}
    Set Suite Variable  ${PROTOCOL}  ${connection_settings[0]}

Cleanup
    Teardown Shared
    Drop Mongodb Database  ${resultDbName}
    Drop Mongodb Database  ${SECOND_MONGO_DB}

Get Request To ${point}
    ${heads}=  Set Variable  ${headers}

    ${resp}=  GET On Session  dbaassession  ${point}

    Log  ${resp}
    Log  ${resp.status_code}
    Log  ${resp.content}

    RETURN  ${resp}

Post Request With ${doc} Data To ${point}
    ${heads}=  Set Variable  ${headers}

    ${resp}=  POST On Session  dbaassession  ${point}  data=${doc}  headers=${heads}

    Log  ${resp}
    Log  ${resp.status_code}
    Log  ${resp.content}

    RETURN  ${resp}

Put Request With ${doc} Data To ${point}
    ${heads}=  Set Variable  ${headers}

    ${resp}=  PUT On Session  dbaassession  ${point}  data=${doc}  headers=${heads}

    Log  ${resp}
    Log  ${resp.status_code}
    Log  ${resp.content}

    RETURN  ${resp}

Wait For ${job} Job Completion With ${attempts} Attempts
    FOR    ${CheckStatus}    IN RANGE    ${attempts}
         ${resp}=  GET On Session  dbaassession  ${job}
         Log  ${resp.content}
         Exit For Loop If    '${resp.status_code}'=='400'
         Exit For Loop If    '${resp.status_code}'=='401'
         Exit For Loop If    '${resp.status_code}'=='403'
         Exit For Loop If    '${resp.status_code}'=='404'
         Exit For Loop If    '${resp.status_code}'=='500'
         ${resultjson}=    Evaluate     json.loads("""${resp.content}""")    json
         Exit For Loop If    '${resultjson['status']}'=='SUCCESS'
         Exit For Loop If    '${resultjson['status']}'=='FAIL'
         Sleep  5s
    END
    Log  ${resp}
    Log  ${resp.content}
    RETURN  ${resp}

Check Meta Values With ${col1} And ${col2} In ${db}
    ${doc}=  Set Variable  {"metadata": {"${col1}":"${col2}"}}
    Check ${doc} Should Exist In ${metaCollection} Of ${db}

Backup Data And Check
    [Arguments]  ${document}  ${attempts}
    ${response}=  Post Request With ${document} Data To /api/${dbaas_api_version}/dbaas/adapter/mongodb/backups/collect
    Should Be Equal As Strings  ${response.status_code}  202
    ${resultjson}=    Evaluate     json.loads("""${response.content}""")    json
    ${backupId}=  Set Variable  ${resultjson['trackId']}
    ${response}=  Wait For /api/${dbaas_api_version}/dbaas/adapter/mongodb/backups/track/backup/${backupId} Job Completion With ${attempts} Attempts
    Should Be Equal As Strings  ${response.status_code}  200
    Set Suite Variable  ${backupId}

Restore Data And Check
    [Arguments]  ${document}  ${backupId}  ${attempts}
    ${response}=  Post Request With ${document} Data To /api/${dbaas_api_version}/dbaas/adapter/mongodb/backups/${backupId}/restore
    Should Be Equal As Strings  ${response.status_code}  202
    ${resultjson}=    Evaluate     json.loads("""${response.content}""")    json
    ${restoreId}=  Set Variable  ${resultjson['trackId']}
    ${response}=  Wait For /api/${dbaas_api_version}/dbaas/adapter/mongodb/backups/track/restore/${restoreId} Job Completion With ${attempts} Attempts
    Should Be Equal As Strings  ${response.status_code}  200

Create Database And Return Users Names
    [Arguments]  ${db_name}
    ${document}=  Set Variable  {"metadata":{"classifier":{"namespace": "${NAMESPACE}"}, "microserviceName":"testMicroservice"},"dbName":"${db_name}"}
    ${resp}=  Post Request With ${document} Data To /api/${dbaas_api_version}/dbaas/adapter/mongodb/databases
    Should Be Equal As Strings  ${resp.status_code}  201
    Dictionary Should Contain Key  ${resp.json()}  name
    ${resp_name}=  Get From Dictionary  ${resp.json()}  name
    Should Be Equal  ${db_name}  ${resp_name}
    Check Roles Existence In Response  ${resp}
    ${resp_con_properties}=  Get From Dictionary  ${resp.json()}  connectionProperties
    ${db_users}=  Get Multi Users Name  ${resp_con_properties}
    RETURN  ${db_users}

Check Roles Existence In Response
    [Arguments]  ${resp}
    ${resp_con_properties}=  Get From Dictionary  ${resp.json()}  connectionProperties
    ${length} =	Get Length	${resp_con_properties}
    Should Be Equal As Integers	 ${length}  4
    Should Contain  str(${resp_con_properties})  'role': 'admin'
    Should Contain  str(${resp_con_properties})  'role': 'streaming'
    Should Contain  str(${resp_con_properties})  'role': 'rw'
    Should Contain  str(${resp_con_properties})  'role': 'ro'

Check Permissions For Role
    [Arguments]  ${db_users}  ${role_name}  ${expected_permissions}
    ${role}=  Get From Dictionary  ${db_users}  ${role_name}
    ${permissions}=  get_permission_for_role  ${MULTI_USERS_DB}  ${role}
    List Should Contain Sub List  ${expected_permissions}  ${permissions}

Check Users Permissions
    [Arguments]  ${db_users}
    ${admin_permissions} =	Create List  dbOwner
    Check Permissions For Role  ${db_users}  admin  ${admin_permissions}
    ${ro_permissions} =	Create List  read
    Check Permissions For Role  ${db_users}  ro  ${ro_permissions}
    ${rw_permissions} =	Create List  readWrite
    Check Permissions For Role  ${db_users}  rw  ${rw_permissions}
    ${streaming_permissions} =	Create List  read  streaming
    Check Permissions For Role  ${db_users}  streaming  ${streaming_permissions}

*** Test Cases ***
Test Wrong Credentials
    [Tags]  mongo  dbaas
    ${wronguser}=  Generate Random String  10
    ${wrongpass}=  Generate Random String  10
    Create Session    wrongcredssession    ${PROTOCOL}://${wronguser}:${wrongpass}@${DBAAS_HOST}:${port}    verify=${verify}
    ${response}=  GET On Session  wrongcredssession  /api/${dbaas_api_version}/dbaas/adapter/mongodb/databases  expected_status=401
    Should Be Equal As Strings  ${response.status_code}  401

Test Create Database
    [Tags]  mongo  dbaas
    ${document}=  Set Variable  {"metadata":{"classifier":{"namespace": "test_namespace"}, "microserviceName":"testMicroservice"},"dbName":"${MONGO_DB}"}
    ${response}=  Post Request With ${document} Data To /api/${dbaas_api_version}/dbaas/adapter/mongodb/databases
    Should Be Equal As Strings  ${response.status_code}  201
    Should Contain  b"${response.content}"  ${resultDbName}
    ${resultjson}=    Evaluate     json.loads("""${response.content}""")    json
    ${users_amount}=  Get length    ${resultjson['connectionProperties']}
    Check Enabled Multi Users
    IF    "${dbaas_api_version}" != "v1" and "${multi_users_enabled}" == "true"
        Should Be Equal As Integers  ${users_amount}  ${4}  Multiple Users Is Enabled, the number of created users is not equal to 4
    ELSE
        Should Be Equal As Integers  ${users_amount}  ${1}  Multiple Users Is Disabled, the number of created users is not equal to 1
    END

Test Check Databases List
    [Tags]  mongo  dbaas
    ${response}=  Get Request To /api/${dbaas_api_version}/dbaas/adapter/mongodb/databases
    Should Be Equal As Strings  ${response.status_code}  200
    Should Contain  b"${response.content}"  ${resultDbName}

Test Update Meta In Database
    [Tags]  mongo  dbaas
    ${updmetacol}=  Generate Random String  10
    ${updmetaval}=  Generate Random String  10

    ${document}=  Set Variable  {"${updmetacol}": "${updmetaval}"}
    ${response}=  Put Request With ${document} Data To /api/${dbaas_api_version}/dbaas/adapter/mongodb/databases/${resultDbName}/metadata
    Log  ${response.content}
    Should Be Equal As Strings  ${response.status_code}  200

    Check Meta Values With ${updmetacol} And ${updmetaval} In ${resultDbName}

Test Create Users To Database
    [Tags]  mongo  dbaas
    ${document}=  Set Variable  {"dbName": "${resultDbName}","password":null}
    ${response}=  Put Request With ${document} Data To /api/${dbaas_api_version}/dbaas/adapter/mongodb/users
    Log  ${response.content}
    Should Be Equal As Strings  ${response.status_code}  201
    ${resultjson}=    Evaluate     json.loads("""${response.content}""")    json

    ${generateduser}=  Set Variable  ${resultjson['connectionProperties']['username']}
    ${generatedpass}=  Set Variable  ${resultjson['connectionProperties']['password']}

    Log  ${generateduser}
    Log  ${generatedpass}

    ${usertocreate}=  Catenate    SEPARATOR=  test   ${generateduser}
    ${passtocreate}=  Catenate    SEPARATOR=  test   ${generatedpass}

    ${document}=  Set Variable  {"metadata":{"classifier":{"namespace": "test_namespace"}, "microserviceName":"testMicroservice"},"dbName": "${resultDbName}","password":"${passtocreate}"}
    ${response}=  Put Request With ${document} Data To /api/${dbaas_api_version}/dbaas/adapter/mongodb/users/${usertocreate}
    Log  ${response.content}
    Should Be Equal As Strings  ${response.status_code}  201
    ${resultjson}=    Evaluate     json.loads("""${response.content}""")    json

    ${predefineduser}=  Set Variable  ${resultjson['connectionProperties']['username']}
    ${predefinedpass}=  Set Variable  ${resultjson['connectionProperties']['password']}

    Log  ${predefineduser}
    Log  ${predefinedpass}

    Should Be Equal As Strings  ${predefineduser}  ${usertocreate}
    Should Be Equal As Strings  ${predefinedpass}  ${passtocreate}

Test Drop Database
    [Tags]  mongo  dbaas
    ${document}=  Set Variable  [{"kind": "database","name": "${resultDbName}"}]
    ${response}=  Post Request With ${document} Data To /api/${dbaas_api_version}/dbaas/adapter/mongodb/resources/bulk-drop
    Should Be Equal As Strings  ${response.status_code}  200

    ${response}=  Get Request To /api/${dbaas_api_version}/dbaas/adapter/mongodb/databases
    Should Be Equal As Strings  ${response.status_code}  200
    Should Not Contain  b"${response.content}"  ${resultDbName}

Test Create Granular Data Directly
    [Tags]  mongo  dbaas backup
    Insert One Mongodb Document  ${collection}  ${documentOne}  database_name=${SECOND_MONGO_DB}
    Insert One Mongodb Document  ${collection}  ${documentTwo}  database_name=${SECOND_MONGO_DB}
    Check ${documentOne} Should Exist In ${collection} Of ${SECOND_MONGO_DB}
    Check ${documentTwo} Should Exist In ${collection} Of ${SECOND_MONGO_DB}

Test Granular Backup
    [Tags]  mongo  dbaas backup
    ${document}=  Set Variable  ["${SECOND_MONGO_DB}"]
    Backup Data And Check  ${document}  ${ATTEMPTS_NUMBER}

Test Delete Granular Data Directly
    [Tags]  mongo  dbaas backup
    Drop Mongodb Database  ${SECOND_MONGO_DB}
    Check ${documentOne} Should Not Exist In ${collection} Of ${SECOND_MONGO_DB}
    Check ${documentTwo} Should Not Exist In ${collection} Of ${SECOND_MONGO_DB}

Test Granular Restore
    [Tags]  mongo  dbaas backup
    ${document}=  Set Variable  ["${SECOND_MONGO_DB}"]
    Restore Data And Check  ${document}  ${backupId}  ${ATTEMPTS_NUMBER}

Test Check Restored Granular Data Directly
    [Tags]  mongo  dbaas backup
    Check ${documentOne} Should Exist In ${collection} Of ${SECOND_MONGO_DB}
    Check ${documentTwo} Should Exist In ${collection} Of ${SECOND_MONGO_DB}

Test Create Data Directly
    [Tags]  mongo  dbaas backup
    Insert One Mongodb Document  ${collection}  ${documentOne}  database_name=${MONGO_DB}
    Insert One Mongodb Document  ${collection}  ${documentTwo}  database_name=${MONGO_DB}
    Check ${documentOne} Should Exist In ${collection} Of ${MONGO_DB}
    Check ${documentTwo} Should Exist In ${collection} Of ${MONGO_DB}

Test Full Backup
    [Tags]  mongo  dbaas backup
    ${document}=  Set Variable  []
    Backup Data And Check  ${document}  ${ATTEMPTS_NUMBER}

Test Delete Data Directly
    [Tags]  mongo  dbaas backup
    Drop Mongodb Database  ${MONGO_DB}
    Check ${documentOne} Should Not Exist In ${collection} Of ${MONGO_DB}
    Check ${documentTwo} Should Not Exist In ${collection} Of ${MONGO_DB}
    Drop Mongodb Database  ${SECOND_MONGO_DB}
    Check ${documentOne} Should Not Exist In ${collection} Of ${SECOND_MONGO_DB}
    Check ${documentTwo} Should Not Exist In ${collection} Of ${SECOND_MONGO_DB}

Test Full Restore
    [Tags]  mongo  dbaas backup
    ${document}=  Set Variable  ["${MONGO_DB}","${SECOND_MONGO_DB}"]
    Restore Data And Check  ${document}  ${backupId}  ${ATTEMPTS_NUMBER}

Test Check All Restored Data Directly
    [Tags]  mongo  dbaas backup
    Check ${documentOne} Should Exist In ${collection} Of ${MONGO_DB}
    Check ${documentTwo} Should Exist In ${collection} Of ${MONGO_DB}
    Check ${documentOne} Should Exist In ${collection} Of ${SECOND_MONGO_DB}
    Check ${documentTwo} Should Exist In ${collection} Of ${SECOND_MONGO_DB}

Test Multi Users Creating
    [Tags]  mongo  dbaas  dbaas_multi_users
    Check Enabled Multi Users
    Skip If  "${dbaas_api_version}" != "v2"  API version v1, not possible to check case!
    Skip If  "${multi_users_enabled}" == "false"  MULTI_USERS_ENABLED = False, not possible to check case!
    ${MULTI_USERS_DB}=  Set Variable  multi_users_${MONGO_DB}
    Set Suite Variable  ${MULTI_USERS_DB}
    ${db_users}=  Create Database And Return Users Names  ${MULTI_USERS_DB}
    Check Users Permissions  ${db_users}
    [Teardown]  Drop Mongodb Database  ${MULTI_USERS_DB}

# Waiting fix for users restore
#Test Users Backup And Restore
#    [Tags]  dbaas  dbaas_multi_users  mongo  dbaas backup
#    Check Enabled Multi Users
#    Skip If  "${dbaas_api_version}" != "v2"  API version v1, not possible to check case!
#    Skip If  "${multi_users_enabled}" == "false"  MULTI_USERS_ENABLED = False, not possible to check case!
#    ${MULTI_USERS_DB}=  Set Variable  multi_restore_${MONGO_DB}
#    Set Suite Variable  ${MULTI_USERS_DB}
#    ${db_users}=  Create Database And Return Users Names  ${MULTI_USERS_DB}
#    Insert One Mongodb Document  ${collection}  ${documentOne}  database_name=${MULTI_USERS_DB}
#    Check ${documentOne} Should Exist In ${collection} Of ${MULTI_USERS_DB}
#    ${all_users}=  Get All Users Names For Db  ${MULTI_USERS_DB}
#    ${db_admin_user_name}=  Get From Dictionary  ${db_users}  admin
#    List Should Contain Value  ${all_users}  ${db_admin_user_name}
#    Check Users Permissions  ${db_users}
#    ${document}=  Set Variable  ["${MULTI_USERS_DB}"]
#    Backup Data And Check  ${document}  ${ATTEMPTS_NUMBER}
#    Delete Db User  ${MULTI_USERS_DB}  ${db_admin_user_name}
#    ${db_streaming_user_name}=  Get From Dictionary  ${db_users}  streaming
#    Revoke Roles From User  ${MULTI_USERS_DB}  ${db_streaming_user_name}  read
#    ${all_users}=  Get All Users Names For Db  ${MULTI_USERS_DB}
#    List Should Not Contain Value  ${all_users}  ${db_admin_user_name}
#    Restore Data And Check  ${document}  ${backupId}  ${ATTEMPTS_NUMBER}
#    ${all_users}=  Get All Users Names For Db  ${MULTI_USERS_DB}
#    List Should Contain Value  ${all_users}  ${db_admin_user_name}
#    Check Users Permissions  ${db_users}
#    [Teardown]  Drop Mongodb Database  ${MULTI_USERS_DB}
