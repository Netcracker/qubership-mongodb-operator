*** Variables ***
${COUNT_OF_RETRY}                   30x
${RETRY_INTERVAL}                   10s
${POST_SCALE_WAIT}                  30s
${DB_DROP_RETRY}                    3x
${DB_DROP_WAIT}                     5s

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
    Separate Datars Hosts

Separate Datars Hosts
    @{list_datars}=  MongoDBLibrary.List Of Datars Hosts
    Set Suite Variable  @{list_datars}

Cleanup
    Run Keyword And Ignore Error  Connect To Mongodb
    Wait Until Keyword Succeeds  ${DB_DROP_RETRY}  ${DB_DROP_WAIT}
    ...  Drop Mongodb Database  ${dbName}

Get Primary And Secondary Replicas
    [Arguments]  ${var}
    ${rs_status_primary}  ${rs_status_secondary} =  get_rs_status  host_repl=${var}
    Log To Console  \nPrimary for ${var}: ${rs_status_primary}
    Log To Console  \nSecondary for ${var}: ${rs_status_secondary}
    [Return]  ${rs_status_primary}  ${rs_status_secondary}

Reelection Primary
    [Arguments]  ${rs_status_primary}  ${var}
    Sleep  ${POST_SCALE_WAIT}
    ${new_rs_status_primary}  ${new_rs_status_secondary}=  Get Primary And Secondary Replicas  ${var}
    ${reelection}=  MongoDBLibrary.Check Reelection Primary  ${rs_status_primary}  ${new_rs_status_primary}
    Log To Console  \nReelection check for ${var}: ${reelection}
    Should Be True  ${reelection}

Fail Replica
    [Arguments]  ${rs_status_primary}  ${obj}
    Sleep  ${POST_SCALE_WAIT}
    ${new_rs_status_primary}  ${new_rs_status_secondary}=  Get Primary And Secondary Replicas  ${obj}
    ${reelection}=  MongoDBLibrary.Check Reelection Primary  ${rs_status_primary}  ${new_rs_status_primary}
    Log To Console  \nReplica failure check for ${obj}: ${reelection}
    Should Not Be True  ${reelection}

MongoDB CRUD Operations
    ${docValueCRUD} =  Generate Random String  10  [LOWER]
    ${docValueCRUDupdated} =  Generate Random String  10  [LOWER]
    ${docCRUD}=  Set Variable  {"name":"${docValueCRUD}"}
    ${docCRUDupd}=  Set Variable  {"name":"${docValueCRUDupdated}"}

    Connect To Mongodb
    Insert One Mongodb Document  ${collection}  ${docCRUD}  database_name=${dbName}
    Check ${docCRUD} Should Exist In ${collection} Of ${dbName}
    Update One Mongodb Document  ${collection}  ${docCRUD}  ${docCRUDupd}  set  database_name=${dbName}
    Check ${docCRUD} Should Not Exist In ${collection} Of ${dbName}
    Check ${docCRUDupd} Should Exist In ${collection} Of ${dbName}
    Delete One Mongodb Document  ${collection}  ${docCRUDupd}  database_name=${dbName}
    Check ${docCRUDupd} Should Not Exist In ${collection} Of ${dbName}

*** Test Cases ***
Test Re-election Primary Datars
    [Tags]  mongo  ha  fail_primary_datars
    [Teardown]  Cleanup
    ${dbName}  ${collection}  ${document}  Create random DB with data
    Set Suite Variable  ${collection}
    Set Suite Variable  ${dbName}

    FOR  ${var}  IN  @{list_datars}
        Log To Console  \n=== Testing re-election for shard ${var} ===
        Connect To Mongodb
        Insert One Mongodb Document  ${collection}  ${document}  database_name=${dbName}
        Check ${document} Should Exist In ${collection} Of ${dbName}
        ${rs_status_primary}  ${rs_status_secondary} =  Get Primary And Secondary Replicas  ${var}
        ${replicas} =  scale_statefulset  ${rs_status_primary}  0
        Wait Until Keyword Succeeds  ${COUNT_OF_RETRY}  ${RETRY_INTERVAL}
        ...  Reelection Primary  ${rs_status_primary}  ${var}

        Sleep  ${POST_SCALE_WAIT}
        Connect To Mongodb
        Check ${document} Should Exist In ${collection} Of ${dbName}
        Insert One Mongodb Document  ${collection}  ${docValueTwo}  database_name=${dbName}
        Check ${docValueTwo} Should Exist In ${collection} Of ${dbName}
        ${replicas} =  scale_statefulset  ${rs_status_primary}  1
        Sleep  ${POST_SCALE_WAIT}
        Connect To Mongodb
        Check ${document} Should Exist In ${collection} Of ${dbName}
        Check ${docValueTwo} Should Exist In ${collection} Of ${dbName}
        Cleanup
    END
    MongoDB CRUD Operations

Test Fail Replica Datars
    [Tags]  mongo  ha  fail_replica_datars
    [Teardown]  Cleanup
    ${dbName}  ${collection}  ${document}  Create random DB with data
    Set Suite Variable  ${collection}
    Set Suite Variable  ${dbName}
    FOR  ${var}  IN  @{list_datars}
        Log To Console  \n=== Testing replica failure for shard ${var} ===
        Connect To Mongodb

        Insert One Mongodb Document  ${collection}  ${document}  database_name=${dbName}
        Check ${document} Should Exist In ${collection} Of ${dbName}
        ${rs_status_primary}  ${rs_status_secondary} =  Get Primary And Secondary Replicas  ${var}
        ${replicas} =  scale_statefulset  ${rs_status_secondary}  0
        Sleep  ${POST_SCALE_WAIT}

        Wait Until Keyword Succeeds  ${COUNT_OF_RETRY}  ${RETRY_INTERVAL}
        ...  Fail Replica  ${rs_status_primary}  ${var}

        Connect To Mongodb
        Check ${document} Should Exist In ${collection} Of ${dbName}
        Insert One Mongodb Document  ${collection}  ${docValueTwo}  database_name=${dbName}
        Check ${docValueTwo} Should Exist In ${collection} Of ${dbName}
        ${replicas} =  scale_statefulset  ${rs_status_secondary}  1
        Sleep  ${POST_SCALE_WAIT}
        Connect To Mongodb
        Check ${document} Should Exist In ${collection} Of ${dbName}
        Check ${docValueTwo} Should Exist In ${collection} Of ${dbName}
        Cleanup
    END
    MongoDB CRUD Operations

Test Fail Primary Cnfrs And Datars
    [Tags]  mongo  ha  fail_primary_datars_cnfrs
    [Teardown]  Cleanup
    ${dbName}  ${collection}  ${document}  Create random DB with data
    Set Suite Variable  ${collection}
    Set Suite Variable  ${dbName}

    Log To Console  \n=== Testing config server + datars primary fail ===
    Connect To Mongodb
    Insert One Mongodb Document  ${collection}  ${document}  database_name=${dbName}
    Check ${document} Should Exist In ${collection} Of ${dbName}
    ${x} =  Get From List  ${list_datars}  0
    Connect To Mongodb Replicas  ${x}
    ${rs_status_datars_primary}  ${rs_status_secondary} =  Get Primary And Secondary Replicas  ${x}
    ${replicas} =  scale_statefulset  ${rs_status_datars_primary}  0
    Wait Until Keyword Succeeds  ${COUNT_OF_RETRY}  ${RETRY_INTERVAL}
    ...  Reelection Primary  ${rs_status_datars_primary}  ${x}
    Connect To Mongodb Replicas  cnfrs
    ${rs_status_cnfrs_primary}  ${rs_status_secondary} =  Get Primary And Secondary Replicas  cnfrs
    ${replicas} =  scale_statefulset  ${rs_status_cnfrs_primary}  0
    Wait Until Keyword Succeeds  ${COUNT_OF_RETRY}  ${RETRY_INTERVAL}
    ...  Reelection Primary  ${rs_status_cnfrs_primary}  cnfrs
    Sleep  ${POST_SCALE_WAIT}
    Connect To Mongodb

    Check ${document} Should Exist In ${collection} Of ${dbName}
    Insert One Mongodb Document  ${collection}  ${docValueTwo}  database_name=${dbName}
    Check ${docValueTwo} Should Exist In ${collection} Of ${dbName}
    ${replicas} =  scale_statefulset  ${rs_status_datars_primary}  1
    ${replicas} =  scale_statefulset  ${rs_status_cnfrs_primary}  1
    Sleep  ${POST_SCALE_WAIT}
    Connect To Mongodb
    Check ${document} Should Exist In ${collection} Of ${dbName}
    Check ${docValueTwo} Should Exist In ${collection} Of ${dbName}
    MongoDB CRUD Operations
