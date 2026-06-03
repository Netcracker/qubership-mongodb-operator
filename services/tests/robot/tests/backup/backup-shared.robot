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

    ${value}=    Get Environment Variable    ${env_name}   default=${EMPTY} 

    IF    not $value
        ${value}=    Get File    ${file_path}
        ${value}=    Strip String    ${value}
    END

    RETURN    ${value}

Load Backup Secrets
    ${backup_user}=    Get Secret Or Env
...    BACKUP_DAEMON_API_CREDENTIALS_USERNAME
...    /var/run/secrets/mongodb/backup-api/username

    ${backup_password}=    Get Secret Or Env
...    BACKUP_DAEMON_API_CREDENTIALS_PASSWORD
...    /var/run/secrets/mongodb/backup-api/password

    ${backup_user}=       Strip String    ${backup_user}
    ${backup_password}=   Strip String    ${backup_password}

    # Set suite variables
    Set Suite Variable    ${BACKUP_DAEMON_API_CREDENTIALS_USERNAME}    ${backup_user}
    Set Suite Variable    ${BACKUP_DAEMON_API_CREDENTIALS_PASSWORD}    ${backup_password}
