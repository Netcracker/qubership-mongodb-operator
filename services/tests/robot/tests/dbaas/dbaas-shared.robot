*** Variables ***
${NAMESPACE}                                      %{OPENSHIFT_WORKSPACE_WA}

*** Settings ***
Library  String
Library	 Collections
Library	 RequestsLibrary
Library  OperatingSystem

*** Keywords ***
Get Secret Or Env
    [Arguments]    ${env_name}    ${file_path}

    ${value}=    Get Environment Variable    ${env_name}    default=${EMPTY}

    IF    not $value
        ${value}=    Get File    ${file_path}
        ${value}=    Strip String    ${value}
    END

    RETURN    ${value}

Load Dbaas Secrets
    ${dbaas_user}=    Get Secret Or Env
...    DBAAS_AGGREGATOR_USERNAME
...    /var/run/secrets/mongodb/dbaas-aggregator/username

    ${dbaas_password}=    Get Secret Or Env
...    DBAAS_AGGREGATOR_PASSWORD
...    /var/run/secrets/mongodb/dbaas-aggregator/password

    ${dbaas_user}=        Strip String    ${dbaas_user}
    ${dbaas_password}=    Strip String    ${dbaas_password}

    Set Suite Variable    ${DBAAS_AGGREGATOR_USERNAME}    ${dbaas_user}
    Set Suite Variable    ${DBAAS_AGGREGATOR_PASSWORD}    ${dbaas_password}

Preparation dbaas shared
    Load Dbaas Secrets
    ${dbaas_api_version}=    Get Dbaas Aggregator version
    Set Suite Variable  ${dbaas_api_version}

Get Dbaas Aggregator version 
    ${env_dbaas_aggregator_host}=  Create List  DBAAS_AGGREGATOR_REGISTRATION_ADDRESS
    ${dbaas_aggregator_host}=  Get Environment Variables For Deployment Entity Container  dbaas-mongo-adapter  ${NAMESPACE}  dbaas-mongo-adapter  ${env_dbaas_aggregator_host}
    Create Session    dbaas_aggregator   ${dbaas_aggregator_host["DBAAS_AGGREGATOR_REGISTRATION_ADDRESS"]}
    Evaluate    logging.getLogger("urllib3").setLevel(logging.ERROR)    logging
    ${resp}=    Run Keyword And Ignore Error    Get On Session    dbaas_aggregator    /api-version
    Evaluate    logging.getLogger("urllib3").setLevel(logging.WARNING)    logging
    Log    ${resp[0]}
    Log    ${resp[1]}

    IF    '${resp[0]}' == 'PASS' 
        IF    '${resp[1].status_code}' == '200'
            ${version}=    Evaluate     json.loads("""${resp[1].content}""")    json
            IF    3 in ${version["supportedMajors"]}
                ${apiVersion}=    Set Variable    v2
            ELSE
                ${apiVersion}=    Set Variable    v1
            END
        END
    ELSE
        ${apiVersion}=    Set Variable    v2
    END
    
    RETURN    ${apiVersion}

Check Enabled Multi Users
    ${env_variables}=  Create List  MULTI_USERS_ENABLED
    ${envs}=  Get Environment Variables For Deployment Entity Container  dbaas-mongo-adapter  ${NAMESPACE}  dbaas-mongo-adapter  ${env_variables}
    ${multi_users_enabled}=  Get From Dictionary  ${envs}  MULTI_USERS_ENABLED
    Set Suite Variable  ${multi_users_enabled}
