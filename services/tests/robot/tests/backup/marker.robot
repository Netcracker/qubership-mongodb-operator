*** Variables ***
${BACKUP_HOST}                                    %{BACKUP_HOST}
${MARKER_VALUE}                                   my-backup/2026-10-07T17:15:00Z


*** Settings ***
Library      String
Library      Collections
Library      RequestsLibrary
Library      OperatingSystem
Resource     backup-shared.robot
Suite Setup  Preparation
Suite Teardown  Delete All Sessions


*** Keywords ***
Preparation
    Load Backup Secrets

    &{headers}=    Create Dictionary    Content-Type=application/json    Accept=application/json
    Set Suite Variable    ${headers}

    ${verify}=      Get Environment Variable    TLS_ROOTCERT    default=False
    ${backup_tls}=  Get Environment Variable    TLS_ENABLED     default=False
    ${port}=        Get Environment Variable    PORT            default=8080

    ${PROTOCOL}=    Set Variable If    '${backup_tls}' == 'true'
    ...    https
    ...    http

    Create Session    backupsession
    ...    ${PROTOCOL}://${BACKUP_DAEMON_API_CREDENTIALS_USERNAME}:${BACKUP_DAEMON_API_CREDENTIALS_PASSWORD}@${BACKUP_HOST}:${port}
    ...    verify=${verify}

Set Marker
    [Arguments]    ${marker}
    ${body}=    Set Variable    {"marker": "${marker}"}
    ${resp}=    POST On Session    backupsession    /marker/set    data=${body}    headers=${headers}
    Log    ${resp.content}
    RETURN    ${resp}

Get Marker
    ${resp}=    GET On Session    backupsession    /marker/get    headers=${headers}
    Log    ${resp.content}
    RETURN    ${resp}


*** Test Cases ***
Test Set Marker Returns 200
    [Tags]    mongo    marker
    ${resp}=    Set Marker    ${MARKER_VALUE}
    Should Be Equal As Strings    ${resp.status_code}    200

Test Get Marker Returns Set Value
    [Tags]    mongo    marker
    ${resp}=    Set Marker    ${MARKER_VALUE}
    Should Be Equal As Strings    ${resp.status_code}    200
    ${resp}=    Get Marker
    Should Be Equal As Strings    ${resp.status_code}    200
    Should Contain    ${resp.text}    ${MARKER_VALUE}

Test Set Marker Overwrites Previous Entry
    [Tags]    mongo    marker
    ${first_marker}=    Set Variable    my-backup/2026-01-01T00:00:00Z
    ${second_marker}=    Set Variable    my-backup/2026-10-07T17:15:00Z
    ${resp}=    Set Marker    ${first_marker}
    Should Be Equal As Strings    ${resp.status_code}    200
    ${resp}=    Set Marker    ${second_marker}
    Should Be Equal As Strings    ${resp.status_code}    200
    ${resp}=    Get Marker
    Should Be Equal As Strings    ${resp.status_code}    200
    Should Contain      ${resp.text}    ${second_marker}
    Should Not Contain  ${resp.text}    ${first_marker}

Test Get Marker After Set Returns Latest Value
    [Tags]    mongo    marker
    ${updated_marker}=    Generate Random String    10    [LOWER]
    ${resp}=    Set Marker    ${updated_marker}
    Should Be Equal As Strings    ${resp.status_code}    200
    ${resp}=    Get Marker
    Should Be Equal As Strings    ${resp.status_code}    200
    Should Contain    ${resp.text}    ${updated_marker}
