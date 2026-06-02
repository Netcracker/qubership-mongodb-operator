*** Variables ***
${NAMESPACE}                                      %{OPENSHIFT_WORKSPACE_WA}

*** Settings ***
Library  String
Library	 Collections
Library	 RequestsLibrary
Library  OperatingSystem

*** Keywords ***
Load Backup Secrets
    ${backup_user}=       Get Environment Variable    BACKUP_DAEMON_API_CREDENTIALS_USERNAME
    ${backup_password}=   Get Environment Variable    BACKUP_DAEMON_API_CREDENTIALS_PASSWORD

    ${backup_user}=       Strip String    ${backup_user}
    ${backup_password}=   Strip String    ${backup_password}

    # Set suite variables
    Set Suite Variable    ${BACKUP_DAEMON_API_CREDENTIALS_USERNAME}    ${backup_user}
    Set Suite Variable    ${BACKUP_DAEMON_API_CREDENTIALS_PASSWORD}    ${backup_password}
