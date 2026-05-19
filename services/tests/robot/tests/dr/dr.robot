*** Variables ***
${COUNT_OF_RETRY}                   10x
${RETRY_INTERVAL}                   5s
${DBAAS_HOST}                       %{DBAAS_HOST}
${BACKUP_HOST}                      %{BACKUP_HOST}
${MAIN_OS_SIDE}                     %{MAIN_OS_SIDE}
${LEFT_NODES_PATTERN}               %{LEFT_NODES_PATTERN}
${RIGHT_NODES_PATTERN}              %{RIGHT_NODES_PATTERN}

*** Settings ***
Library  String
Library  OperatingSystem
Resource  ../shared/keywords.robot
Suite Setup  Preparation
Suite Teardown  Cleanup

*** Keywords ***
Preparation
    Prepare Shared
    ${docValueRand}=  Generate Random String  10  [LOWER]
    ${docValueTwo}=  Set Variable  {"name":"${docValueRand}new"}
    Set Suite Variable  ${docValueTwo}
    ${dbName}=  Generate Random String  10  [LOWER]
    Set Suite Variable  ${dbName}
    ${collection}=  Generate Random String  10  [LOWER]
    Set Suite Variable  ${collection}
    Create Hardcoded DR Test Data
    Cleanup Hardcoded Data
    Separate Datars Hosts
    @{components_single}=  Create List  mongos
    Run Keyword If  '${DBAAS_HOST}' != '${EMPTY}'  Append To List  ${components_single}  dbaas-mongo-adapter
    Run Keyword If  '${BACKUP_HOST}' != '${EMPTY}'  Append To List  ${components_single}  mongodb-backup-daemon
    Set Suite Variable  @{components_single}
    @{components_plural}=  Create List  cnfrs  @{list_datars}
    Set Suite Variable  @{components_plural}
    ${first_datars} =  Get From List  ${list_datars}  0
    Set Suite Variable  ${first_datars}

Cleanup
    Drop Mongodb Database  ${dbName}

Cleanup Hardcoded Data
    Drop Mongodb Database  ${hardcoded_dbName}

Separate Datars Hosts
    @{list_datars}=  MongoDBLibrary.List Of Datars Hosts
    Set Suite Variable  @{list_datars}

Create Hardcoded DR Test Data
    ${hardcoded_dbName} =  Set Variable  test_dr_db
    ${hardcoded_collection} =  Set Variable  test_dr_collection
    ${hardcoded_test_value} =  Set Variable  test_dr_value
    ${hardcoded_document}=  Set Variable  {"name":"${hardcoded_test_value}"}
    Set Suite Variable  ${hardcoded_dbName}
    Set Suite Variable  ${hardcoded_collection}
    Set Suite Variable  ${hardcoded_document}

MongoDB CRUD
    ${docValueCRUD} =  Generate Random String  10  [LOWER]
    ${docValueCRUDupdated} =  Generate Random String  10  [LOWER]
    ${docCRUD}=  Set Variable  {"name":"${docValueCRUD}"}
    ${docCRUDupd}=  Set Variable  {"name":"${docValueCRUDupdated}"}
    Insert One Mongodb Document  ${collection}  ${docCRUD}  database_name=${dbName}
    Check ${docCRUD} Should Exist In ${collection} Of ${dbName}
    Update One Mongodb Document  ${collection}  ${docCRUD}  ${docCRUDupd}  set  database_name=${dbName}
    Check ${docCRUD} Should Not Exist In ${collection} Of ${dbName}
    Check ${docCRUDupd} Should Exist In ${collection} Of ${dbName}
    Delete One Mongodb Document  ${collection}  ${docCRUDupd}  database_name=${dbName}
    Check ${docCRUDupd} Should Not Exist In ${collection} Of ${dbName}

Check Status Of Replication
    [Arguments]  ${scheme}
    FOR  ${var}  IN  @{list_datars}
        ${resp} =  check_replication  ${scheme}  host_repl=${var}
        Should Be True  ${resp}
    END

Check Location Of Component
    [Arguments]  ${service}  ${mode}
    ${left_pods}  ${right_pods} =  Get Side For Pod  ${NAMESPACE}  ${service}  ${LEFT_NODES_PATTERN}  ${RIGHT_NODES_PATTERN}
    ${resp} =  Check Location Component  main_os=${MAIN_OS_SIDE}  left_pods=${left_pods}  right_pods=${right_pods}  mode=${mode}
    Should Be True  ${resp}

Check Location Single Components
    FOR  ${var}  IN  @{components_single}
        Check Location Of Component  ${var}  single
    END

Check Location Plural Components
    [Arguments]  ${mode}
    FOR  ${var}  IN  @{components_plural}
        Check Location Of Component  ${var}  ${mode}
    END


*** Test Cases ***
Test Check MongoDB Before Switchover
    [Tags]  dr  dr_before_switchover
    MongoDB CRUD
    Check Status Of Replication  not_failover
    Insert One Mongodb Document  ${hardcoded_collection}  ${hardcoded_document}  database_name=${hardcoded_dbName}
    Check ${hardcoded_document} Should Exist In ${hardcoded_collection} Of ${hardcoded_dbName}
    Check Location Single Components
    Check Location Plural Components  plural

Test Check MongoDB After Switchover
    [Tags]  dr  dr_after_switchover
    [Teardown]  Cleanup Hardcoded Data
    Check Status Of Replication  not_failover
    MongoDB CRUD
    Check ${hardcoded_document} Should Exist In ${hardcoded_collection} Of ${hardcoded_dbName}
    Check Location Single Components
    Check Location Plural Components  plural

Test Check MongoDB Before Failover
    [Tags]  dr  dr_before_failover
    MongoDB CRUD
    Check Status Of Replication  not_failover
    Insert One Mongodb Document  ${hardcoded_collection}  ${hardcoded_document}  database_name=${hardcoded_dbName}
    Check ${hardcoded_document} Should Exist In ${hardcoded_collection} Of ${hardcoded_dbName}
    Check Location Single Components
    Check Location Plural Components  plural

Test Check MongoDB After Failover
    [Tags]  dr  dr_after_failover
    Check Status Of Replication  failover
    MongoDB CRUD
    Check ${hardcoded_document} Should Exist In ${hardcoded_collection} Of ${hardcoded_dbName}
    Check Location Single Components
    Check Location Plural Components  plural_failover

Test Check MongoDB After Return
    [Tags]  dr  dr_after_return
    [Teardown]  Cleanup Hardcoded Data
    Check Status Of Replication  not_failover
    MongoDB CRUD
    Check ${hardcoded_document} Should Exist In ${hardcoded_collection} Of ${hardcoded_dbName}
    Check Location Single Components
    Check Location Plural Components  plural
