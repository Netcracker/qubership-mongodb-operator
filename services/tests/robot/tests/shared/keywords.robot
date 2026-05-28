*** Variables ***
${MONGO_HOST}                  %{MONGO_HOST}
${WAIT_TIMEOUT}                %{WAIT_TIMEOUT}
${NAMESPACE}                   %{OPENSHIFT_WORKSPACE_WA}
${DATARS_HOST}                 %{DATARS_HOST}
${TLS_ENABLED}                 %{TLS_ENABLED=False}
${TLS_ROOTCERT}                %{TLS_ROOTCERT=None}

*** Settings ***
Library  String
Library  Collections
Library  RequestsLibrary
Library  ../lib/KubernetesClient.py

*** Keywords ***
Load Secrets
    ${mongo_user}=        Get Environment Variable    MONGO_ROOT_USER
    ${mongo_password}=    Get Environment Variable    MONGO_ROOT_PASSWORD

    ${mongo_user}=        Strip String    ${mongo_user}
    ${mongo_password}=    Strip String    ${mongo_password}

    # Set suite variables
    Set Suite Variable    ${MONGO_ROOT_USER}              ${mongo_user}
    Set Suite Variable    ${MONGO_ROOT_PASSWORD}          ${mongo_password}

Prepare Shared
    Load Secrets

    &{headers}=    Create Dictionary    Content-Type=application/json    Accept=application/json
    Set Suite Variable    ${headers}

    ${MONGO_DB}=    Generate Random String    10    [LOWER]
    Set Suite Variable    ${MONGO_DB}

    Import Library    ${CURDIR}/../lib/MongoDBLibrary.py
    ...    host=${MONGO_HOST}
    ...    port=27017
    ...    user=${MONGO_ROOT_USER}
    ...    password=${MONGO_ROOT_PASSWORD}
    ...    database_name=test
    ...    req_timeout_sec=${WAIT_TIMEOUT}
    ...    host_datars=${DATARS_HOST}
    ...    tls=${TLS_ENABLED}
    ...    tlsCAFile=${TLS_ROOTCERT}

    Test Mongo Connection

Teardown Shared
    Drop Mongodb Database  ${MONGO_DB}

Test Mongo Connection
    Connect To Mongodb

Check ${var} environment variable on ${pod} is ${value}
    ${val}=  Get Var On Pod  ${var}  ${pod}
    ${res}=  set variable  '${val}'=='${value}'
    RETURN  ${res}

Check ${document} Should Exist In ${collection} Of ${db}
    ${result}=  Check Document Exists  ${collection}  ${document}  database_name=${db}
    Log  ${result}
    Should Be True  ${result}

Check ${document} Should Not Exist In ${collection} Of ${db}
    ${result}=  Check Document Exists  ${collection}  ${document}  database_name=${db}
    Log  ${result}
    Should Not Be True  ${result}

Create random DB with data
    ${dbName}=  Generate Random String  10  [LOWER]

    ${docValue}=  Generate Random String  10  [LOWER]
    ${document}=  Set Variable  {"name":"${docValue}"}

    ${collection}=  Generate Random String  10  [LOWER]

    RETURN  ${dbName}  ${collection}  ${document}
