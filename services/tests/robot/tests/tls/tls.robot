*** Variables ***
${DBAAS_HOST}                                     %{DBAAS_HOST}
${DBAAS_AGGREGATOR_USERNAME}                      %{DBAAS_AGGREGATOR_USERNAME}
${DBAAS_AGGREGATOR_PASSWORD}                      %{DBAAS_AGGREGATOR_PASSWORD}
${BACKUP_HOST}                                    %{BACKUP_HOST}
${BACKUP_DAEMON_API_CREDENTIALS_USERNAME}         %{BACKUP_DAEMON_API_CREDENTIALS_USERNAME}
${BACKUP_DAEMON_API_CREDENTIALS_PASSWORD}         %{BACKUP_DAEMON_API_CREDENTIALS_PASSWORD}
${TLS_ENABLED}                                    %{TLS_ENABLED=false}
${NAMESPACE}                                      %{OPENSHIFT_WORKSPACE_WA}

*** Settings ***
Library  String
Library	 Collections
Library  OperatingSystem
Library  RequestsLibrary
Library  ../lib/MongoDBLibrary.py  host=%{MONGO_HOST}
...                                port=27017
...                                user=%{MONGO_ROOT_USER}
...                                password=%{MONGO_ROOT_PASSWORD}
...                                database_name=test_tls
...                                req_timeout_sec=%{WAIT_TIMEOUT}
...                                host_datars=%{DATARS_HOST}
...                                tls=${False}
...                                tlsCAFile=${None}
Library  ../lib/KubernetesClient.py
Suite Setup  Check HTTPS Enabling in Dbaas Aggregator

*** Keywords ***
Check HTTPS Enabling in Dbaas Aggregator
    ${env_dbaas_aggregator_host}=  Create List  DBAAS_AGGREGATOR_REGISTRATION_ADDRESS
    ${dbaas_aggregator_host}=  Get Environment Variables For Deployment Entity Container  dbaas-mongo-adapter  ${NAMESPACE}  dbaas-mongo-adapter  ${env_dbaas_aggregator_host}
    ${https_aggregator_enabled}=  Evaluate  "https" in "${dbaas_aggregator_host}"
    Set Suite Variable  ${https_aggregator_enabled}
    ${verify}=  Get Environment Variable  name=TLS_ROOTCERT  default=False
    ${port}=  Get Environment Variable  name=PORT  default=8080
    Set Suite Variable  ${verify}
    Set Suite Variable  ${port}

*** Test Cases ***
Check Connection To TLS MongoDB Without Cert
    [Tags]  tls  mongo
    Skip If  "${TLS_ENABLED}" == "false"  TLS_ENABLED = False, not possible to check case!
    Connect To Mongodb
    Run Keyword And Expect Error  *  Insert One Mongodb Document  robot_collection  {"name" : "Tom"}  database_name=test_tls

Check Connection To TLS Dbaas Adapter By HTTP Protocol
    [Tags]  tls  mongo
    Check HTTPS Enabling in Dbaas Aggregator
    Skip If  "${TLS_ENABLED}" == "false"  TLS_ENABLED = False, not possible to check case!
    Skip If  "${https_aggregator_enabled}" == "${False}"  HTTP is already active protocol!
    Create Session  wrongprotocolsession  http://${DBAAS_AGGREGATOR_USERNAME}:${DBAAS_AGGREGATOR_PASSWORD}@${DBAAS_HOST}:${port}  verify=${verify}  timeout=10
    Run Keyword And Expect Error  *ProtocolError*  GET On Session  wrongprotocolsession  /api/${dbaas_api_version}/dbaas/adapter/mongodb/databases

Check Connection To TLS Dbaas Adapter By Not Correct Port
    [Tags]  tls  mongo
    Skip If  "${TLS_ENABLED}" == "false"  TLS_ENABLED = False, not possible to check case!
    Create Session  wrongportsession  https://${DBAAS_AGGREGATOR_USERNAME}:${DBAAS_AGGREGATOR_PASSWORD}@${DBAAS_HOST}:8080  verify=${verify}  timeout=10
    Run Keyword And Expect Error  *  GET On Session  wrongportsession  /api/${dbaas_api_version}/dbaas/adapter/mongodb/databases

Check Connection To TLS Backup Daemon By HTTP Protocol
    [Tags]  tls  mongo
    Skip If  "${TLS_ENABLED}" == "false"  TLS_ENABLED = False, not possible to check case!
    Create Session  wrongprotocolsession  http://${BACKUP_DAEMON_API_CREDENTIALS_USERNAME}:${BACKUP_DAEMON_API_CREDENTIALS_PASSWORD}@${BACKUP_HOST}:${port}  verify=${verify}  timeout=10
    Run Keyword And Expect Error  *HTTPError*  GET On Session  wrongprotocolsession  /backup

Check Connection To TLS Backup Daemon By Not Correct Port
    [Tags]  tls  mongo
    Skip If  "${TLS_ENABLED}" == "false"  TLS_ENABLED = False, not possible to check case!
    Create Session  wrongportsession  https://${BACKUP_DAEMON_API_CREDENTIALS_USERNAME}:${BACKUP_DAEMON_API_CREDENTIALS_PASSWORD}@${BACKUP_HOST}:8080  verify=${verify}  timeout=10
    Run Keyword And Expect Error  *  GET On Session  wrongportsession  /backup