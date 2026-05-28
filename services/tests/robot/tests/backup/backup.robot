*** Variables ***
${BACKUP_HOST}                                    %{BACKUP_HOST}
${EXTERNAL_BACKUP_PATH}                           %{EXTERNAL_BACKUP_PATH}


*** Settings ***
Library  String
Library	 Collections
Library	 RequestsLibrary
Resource  ../shared/keywords.robot
Resource  backup-shared.robot
Library  ../lib/KubernetesClient.py
Library    OperatingSystem
Suite Setup  Preparation
Suite Teardown  Cleanup

*** Keywords ***
Preparation
    Prepare Shared
    Load Backup Secrets

    ${full_restore}=  Check ENABLE_FULL_RESTORE environment variable on name=mongodb-backup-daemon is true
    Set Suite Variable  ${full_restore}

    ${backup_mount}=  get external mounts  service=mongodb-backup-daemon  label=name
    Set Suite Variable  ${backup_mount}

    ${SECOND_MONGO_DB}=  Set Variable  granular_${MONGO_DB}
    Set Suite Variable  ${SECOND_MONGO_DB}
    ${NFS_MONGO_DB}=  Set Variable  granular_nfs_${MONGO_DB}
    Set Suite Variable  ${NFS_MONGO_DB}

    ${collection}=  Generate Random String  10  [LOWER]
    Set Suite Variable  ${collection}

    ${docValueOne}=  Generate Random String  10  [LOWER]
    ${docValueTwo}=  Generate Random String  10  [LOWER]

    ${documentOne}=  Set Variable  {"name":"${docValueOne}"}
    ${documentTwo}=  Set Variable  {"name":"${docValueTwo}"}
    Set Suite Variable  ${documentOne}
    Set Suite Variable  ${documentTwo}

    ${skipMessage}=  Set Variable  Skipped because ENABLE_FULL_RESTORE is disabled
    Set Suite Variable  ${skipMessage}
    ${skipMessageNFS}=  Set Variable  Skipped because mongodb-backup-daemon was installed without specified external NFS
    Set Suite Variable  ${skipMessageNFS}

    ${verify}=    Get Environment Variable    name=TLS_ROOTCERT    default=False
    ${backup_tls}=    Get Environment Variable    name=TLS_ENABLED    default=False
    ${port}=    Get Environment Variable    name=PORT    default=8080

    ${PROTOCOL} =    Set Variable If    '${backup_tls}' == 'true'
        ...  https
        ...  http

    Create Session    backupsession    ${PROTOCOL}://${BACKUP_DAEMON_API_CREDENTIALS_USERNAME}:${BACKUP_DAEMON_API_CREDENTIALS_PASSWORD}@${BACKUP_HOST}:${port}   verify=${verify}

Cleanup
    Teardown Shared
    Drop Mongodb Database  ${dbNameFull}
    Drop Mongodb Database  ${NFS_MONGO_DB}
    run keyword if  ${full_restore}  Drop Mongodb Database  ${dbNameInc}
    run keyword if  ${full_restore}  Drop Mongodb Database  ${dbNameNFS}
    run keyword if  ${full_restore}  Drop Mongodb Database  ${dbNameIncrementalNFS}

Wait For ${job} Job Completion With ${attempts} Attempts On ${endpoint}
    FOR    ${CheckStatus}    IN RANGE    ${attempts}
         ${resp}=  GET On Session  backupsession  ${endpoint}/jobstatus/${job}
         Log  ${resp.content}
         Exit For Loop If    '${resp.status_code}'=='200'
         Exit For Loop If    '${resp.status_code}'=='400'
         Exit For Loop If    '${resp.status_code}'=='401'
         Exit For Loop If    '${resp.status_code}'=='500'
         Sleep  5s
    END
    Log  ${resp}
    Log  ${resp.content}
    [Return]  ${resp}

Run ${doc} Job To ${point} Endpoint And Check With ${attempts} Attempts On ${endpoint}
    ${heads}=  Set Variable  ${None}
    ${heads}=  Run Keyword If  ${doc}  Set Variable  ${headers}

    ${resp}=  POST On Session  backupsession  ${point}  data=${doc}  headers=${heads}
    Log  ${resp}
    Log  ${resp.content}
    Should Be Equal As Strings  ${resp.status_code}  200
    ${currbackupjob}=  Set Variable  ${resp.content}

    ${resp}=  Wait For ${currbackupjob} Job Completion With ${attempts} Attempts On ${endpoint}
    Log  ${resp.content}
    Should Be Equal As Strings  ${resp.status_code}  200

    [Return]  ${currbackupjob}

Backup ${doc} Data And Check With ${attempts} Attempts On ${endpoint}
    ${job}=  Run ${doc} Job To ${endpoint}/backup Endpoint And Check With ${attempts} Attempts On ${endpoint}
    [Return]  ${job}

Restore ${doc} Data And Check With ${attempts} Attempts On ${endpoint}
    ${job}=  Run ${doc} Job To ${endpoint}/restore Endpoint And Check With ${attempts} Attempts On ${endpoint}

Find Name of Backup Pod
    ${backup_pod}=  Get Pod For Service  namespace=${NAMESPACE}    service=mongodb-backup-daemon    label_name=name
    Set Suite Variable  ${backup_pod}

Check Location Of Backup
    [Arguments]  ${PATH_TO_BACKUP}  ${BACKUP_ID}
    Find Name of Backup Pod
    ${list_external_backups}=  Execute Command In Pod  ${backup_pod}  ${NAMESPACE}  ls /external/${PATH_TO_BACKUP}/
    ${l}=  convert to string  ${list_external_backups}
    ${id}=  convert to string  ${BACKUP_ID}
    Should Contain  ${l}  ${id}
    ${nfs_backup_folder}=  Execute Command In Pod  ${backup_pod}  ${NAMESPACE}  ls /external/${PATH_TO_BACKUP}/${BACKUP_ID}/
    ${backup_folder}=  convert to string  ${nfs_backup_folder}
    Should Not Be Empty  ${backup_folder}

Check Duplicates Of Backups
    [Arguments]  ${PATH}  ${BACKUP_ID}
    Find Name of Backup Pod
    ${list}=  Execute Command In Pod    ${backup_pod}  ${NAMESPACE}  ls ${PATH}
    ${l1}=  convert to string  ${list}
    ${id1}=  convert to string  ${BACKUP_ID}
    Should Not Contain  ${l1}  ${id1}

*** Test Cases ***
Test Create Granular Data
    [Tags]  mongo  backup
    Insert One Mongodb Document  ${collection}  ${documentOne}  database_name=${MONGO_DB}
    Insert One Mongodb Document  ${collection}  ${documentTwo}  database_name=${MONGO_DB}
    Check ${documentOne} Should Exist In ${collection} Of ${MONGO_DB}
    Check ${documentTwo} Should Exist In ${collection} Of ${MONGO_DB}

Test Backup Granular Data
    [Tags]  mongo  backup
    ${document}=  Set Variable  {"dbs":["${MONGO_DB}"], "allow_eviction":"true"}
    ${granularbackupjob}=  Backup ${document} Data And Check With 100 Attempts On ${Empty}
    Set Suite Variable  ${granularbackupjob}

Test Delete Granular Data
    [Tags]  mongo  backup
    Drop Mongodb Database  ${MONGO_DB}
    Check ${documentOne} Should Not Exist In ${collection} Of ${MONGO_DB}
    Check ${documentTwo} Should Not Exist In ${collection} Of ${MONGO_DB}

Test Restore Granular Data
    [Tags]  mongo  backup
    ${document}=  Set Variable  {"vault":"${granularbackupjob}","dbs":["${MONGO_DB}"]}
    Restore ${document} Data And Check With 100 Attempts On ${Empty}

Test Check Restored Granular Data
    [Tags]  mongo  backup
    Check ${documentOne} Should Exist In ${collection} Of ${MONGO_DB}
    Check ${documentTwo} Should Exist In ${collection} Of ${MONGO_DB}

Test Create Data
    [Tags]  mongo  backup
    ${dbNameFull}  ${collectionNameFull}  ${documentFull}  Create random DB with data
    Insert One Mongodb Document  ${collectionNameFull}  ${documentFull}   database_name=${dbNameFull}
    Check ${documentFull} Should Exist In ${collectionNameFull} Of ${dbNameFull}
    Set Suite Variable  ${dbNameFull}
    Set Suite Variable  ${collectionNameFull}
    Set Suite Variable  ${documentFull}

Test Full Backup Data
    [Tags]  mongo  backup
    ${document}=  Set Variable  {"allow_eviction":"true"}
    ${backupjob}=  Backup ${document} Data And Check With 100 Attempts On ${Empty}
    Set Suite Variable  ${backupjob}

Test Create Incremental Data
    [Tags]  mongo  externalbackup
    Pass Execution If  not ${full_restore}  ${skipMessage}
    ${dbNameInc}  ${collectionNameInc}  ${documentInc}  Create random DB with data
    Insert One Mongodb Document  ${collectionNameInc}  ${documentInc}   database_name=${dbNameInc}
    Check ${documentInc} Should Exist In ${collectionNameInc} Of ${dbNameInc}
    Set Suite Variable  ${dbNameInc}
    Set Suite Variable  ${collectionNameInc}
    Set Suite Variable  ${documentInc}

Test Backup Incremental Data
    [Tags]  mongo  externalbackup
    Pass Execution If  not ${full_restore}  ${skipMessage}
    ${endpoint}=  Set Variable  /incremental
    ${incbackupjob}=  Backup ${None} Data And Check With 100 Attempts On ${endpoint}
    Set Suite Variable  ${incbackupjob}

Test Delete Incremental Data
    [Tags]  mongo  externalbackup
    Pass Execution If  not ${full_restore}  ${skipMessage}
    Drop Mongodb Database  ${dbNameInc}
    Check ${documentInc} Should Not Exist In ${collectionNameInc} Of ${dbNameInc}

Test Restore Incremental Data
    [Tags]  mongo  externalbackup
    Pass Execution If  not ${full_restore}  ${skipMessage}
    ${document}=  Set Variable  {"vault":"${incbackupjob}"}
    ${endpoint}=  Set Variable  /incremental
    Restore ${document} Data And Check With 100 Attempts On ${endpoint}

Test Check Restored Incremental Data
    [Tags]  mongo  externalbackup
    Pass Execution If  not ${full_restore}  ${skipMessage}
    Check ${documentInc} Should Exist In ${collectionNameInc} Of ${dbNameInc}

Test Delete Data
    [Tags]  mongo  backup
    Drop Mongodb Database  ${dbNameFull}
    Check ${documentFull} Should Not Exist In ${collectionNameFull} Of ${dbNameFull}

Test Restore Data
    [Tags]  mongo  backup
    ${document}=  Set Variable  {"vault":"${backupjob}","dbs":["${dbNameFull}"]}
    Restore ${document} Data And Check With 100 Attempts On ${Empty}

Test Check All Restored Data
    [Tags]  mongo  backup
    Check ${documentFull} Should Exist In ${collectionNameFull} Of ${dbNameFull}

#Granular NFS
Test Create Granular Data For Backup On External Storage
    [Tags]  mongo  externalbackup
    Pass Execution If  not ${backup_mount}  ${skipMessageNFS}
    Insert One Mongodb Document  ${collection}  ${documentOne}  database_name=${NFS_MONGO_DB}
    Insert One Mongodb Document  ${collection}  ${documentTwo}  database_name=${NFS_MONGO_DB}
    Log to console  database_name=${NFS_MONGO_DB}
    Check ${documentOne} Should Exist In ${collection} Of ${NFS_MONGO_DB}
    Check ${documentTwo} Should Exist In ${collection} Of ${NFS_MONGO_DB}

Test Backup Granular Data On External NFS Storage
    [Tags]  mongo  externalbackup
    Pass Execution If  not ${backup_mount}  ${skipMessageNFS}
    ${document}=  Set Variable  {"externalBackupPath":"${EXTERNAL_BACKUP_PATH}","dbs":["${NFS_MONGO_DB}"]}
    ${granularnfsbackupjob}=  Backup ${document} Data And Check With 100 Attempts On ${Empty}
    Set Suite Variable  ${granularnfsbackupjob}
    Log to console  granularnfsbackupjob: ${granularnfsbackupjob}

Test Backup Located On External NFS Storage
    [Tags]  mongo  externalbackup
    Pass Execution If  not ${backup_mount}  ${skipMessageNFS}
    Check Location Of Backup  ${EXTERNAL_BACKUP_PATH}  ${granularnfsbackupjob}
    Check Duplicates Of Backups  /backup-storage/granular/  ${granularnfsbackupjob}

Test Delete Granular Data For Backup On External NFS Storage
    [Tags]  mongo  externalbackup
    Pass Execution If  not ${backup_mount}  ${skipMessageNFS}
    Drop Mongodb Database  ${NFS_MONGO_DB}
    Check ${documentOne} Should Not Exist In ${collection} Of ${NFS_MONGO_DB}
    Check ${documentTwo} Should Not Exist In ${collection} Of ${NFS_MONGO_DB}

Test Restore Data From Backup On External NFS Storage
    [Tags]  mongo  externalbackup
    Pass Execution If  not ${backup_mount}  ${skipMessageNFS}
    ${document}=  Set Variable  {"externalBackupPath":"${EXTERNAL_BACKUP_PATH}","vault":"${granularnfsbackupjob}","dbs":["${NFS_MONGO_DB}"]}
    Restore ${document} Data And Check With 100 Attempts On ${Empty}

Test Check All Restored Granular Data
    [Tags]  mongo  externalbackup
    Pass Execution If  not ${backup_mount}  ${skipMessageNFS}
    Check ${documentOne} Should Exist In ${collection} Of ${NFS_MONGO_DB}
    Check ${documentTwo} Should Exist In ${collection} Of ${NFS_MONGO_DB}

#Full NFS
Test Create Data For Full Backup On NFS Storage
    [Tags]  mongo  externalbackup
    Pass Execution If  not ${backup_mount}  ${skipMessageNFS}
    Pass Execution If  not ${full_restore}  ${skipMessage}
    ${dbNameFullNFS}  ${collectionNameFullNFS}  ${documentFullNFS}  Create random DB with data
    Insert One Mongodb Document  ${collectionNameFullNFS}  ${documentFullNFS}   database_name=${dbNameFullNFS}
    Check ${documentFullNFS} Should Exist In ${collectionNameFullNFS} Of ${dbNameFullNFS}
    Set Suite Variable  ${dbNameFullNFS}
    Set Suite Variable  ${collectionNameFullNFS}
    Set Suite Variable  ${documentFullNFS}

Test Full Backup On External NFS Storage
    [Tags]  mongo  externalbackup
    Pass Execution If  not ${backup_mount}  ${skipMessageNFS}
    Pass Execution If  not ${full_restore}  ${skipMessage}
    ${document}=  Set Variable  {"externalBackupPath":"${EXTERNAL_BACKUP_PATH}"}
    ${fullnfsbackupjob}=  Backup ${document} Data And Check With 100 Attempts On ${Empty}
    Log to console  Fullnfsbackupjob:${fullnfsbackupjob}
    Set Suite Variable  ${fullnfsbackupjob}

Test Full Backup Located On External NFS Storage
    [Tags]  mongo  externalbackup
    Pass Execution If  not ${backup_mount}  ${skipMessageNFS}
    Check Location Of Backup  ${EXTERNAL_BACKUP_PATH}  ${fullnfsbackupjob}
    Check Duplicates Of Backups  /backup-storage/  ${fullnfsbackupjob}

Test Delete Full Data For Backup On External NFS Storage
    [Tags]  mongo  externalbackup
    Pass Execution If  not ${backup_mount}  ${skipMessageNFS}
    Pass Execution If  not ${full_restore}  ${skipMessage}
    Drop Mongodb Database  ${dbNameFullNFS}
    Check ${documentFullNFS} Should Not Exist In ${collectionNameFullNFS} Of ${dbNameFullNFS}

Test Restore Data From Full Backup On External NFS Storage
    [Tags]  mongo  externalbackup
    Pass Execution If  not ${backup_mount}  ${skipMessageNFS}
    Pass Execution If  not ${full_restore}  ${skipMessage}
    ${document}=  Set Variable  {"externalBackupPath":"${EXTERNAL_BACKUP_PATH}","vault":"${fullnfsbackupjob}","dbs":["${dbNameFullNFS}"]}
    Restore ${document} Data And Check With 100 Attempts On ${Empty}

Test Check All Restored Full Data
    [Tags]  mongo  externalbackup
    Pass Execution If  not ${backup_mount}  ${skipMessageNFS}
    Pass Execution If  not ${full_restore}  ${skipMessage}
    Check ${documentFullNFS} Should Exist In ${collectionNameFullNFS} Of ${dbNameFullNFS}
    Drop Mongodb Database  ${dbNameFullNFS}

#Incremental NFS
Test Create Data For Full Backup Before Incremental On NFS Storage
    [Tags]  mongo  externalbackup
    Pass Execution If  not ${backup_mount}  ${skipMessageNFS}
    Pass Execution If  not ${full_restore}  ${skipMessage}
    ${dbNameNFS}  ${collectionNameNFS}  ${documentNFS}  Create random DB with data
    Insert One Mongodb Document  ${collectionNameNFS}  ${documentNFS}   database_name=${dbNameNFS}
    Check ${documentNFS} Should Exist In ${collectionNameNFS} Of ${dbNameNFS}
    Set Suite Variable  ${dbNameNFS}
    Set Suite Variable  ${collectionNameNFS}
    Set Suite Variable  ${documentNFS}
    Log to console  dbNameNFS:${dbNameNFS}

Test Full Backup Before Incremental On External NFS Storage
    [Tags]  mongo  externalbackup
    Pass Execution If  not ${backup_mount}  ${skipMessageNFS}
    Pass Execution If  not ${full_restore}  ${skipMessage}
    ${document}=  Set Variable  {"externalBackupPath":"${EXTERNAL_BACKUP_PATH}"}
    ${fullnfsbackupjob}=  Backup ${document} Data And Check With 100 Attempts On ${Empty}
    Log to console  Fullnfsbackupjob:${fullnfsbackupjob}
    Set Suite Variable  ${fullnfsbackupjob}

Test Create Data For Incremental Backup On NFS Storage
    [Tags]  mongo  externalbackup
    Pass Execution If  not ${backup_mount}  ${skipMessageNFS}
    Pass Execution If  not ${full_restore}  ${skipMessage}
    ${dbNameIncrementalNFS}  ${collectionNameIncrementalNFS}  ${documentIncrementalNFS}  Create random DB with data
    Insert One Mongodb Document  ${collectionNameIncrementalNFS}  ${documentIncrementalNFS}   database_name=${dbNameIncrementalNFS}
    Check ${documentIncrementalNFS} Should Exist In ${collectionNameIncrementalNFS} Of ${dbNameIncrementalNFS}
    Set Suite Variable  ${dbNameIncrementalNFS}
    Set Suite Variable  ${collectionNameIncrementalNFS}
    Set Suite Variable  ${documentIncrementalNFS}
    Log to console  dbNameIncrementalNFS:${dbNameIncrementalNFS}

Test Incremental Backup Data On External NFS Storage
    [Tags]  mongo  externalbackup
    Pass Execution If  not ${backup_mount}  ${skipMessageNFS}
    Pass Execution If  not ${full_restore}  ${skipMessage}
    ${document}=  Set Variable  {"externalBackupPath":"${EXTERNAL_BACKUP_PATH}"}
    ${endpoint}=  Set Variable  /incremental
    ${incnfsbackupjob}=  Backup ${document} Data And Check With 100 Attempts On ${endpoint}
    Log to console  Incnfsbackupjob:${incnfsbackupjob}
    Set Suite Variable  ${incnfsbackupjob}

Test Incremental Backup Located On External NFS Storage
    [Tags]  mongo  externalbackup
    Pass Execution If  not ${backup_mount}  ${skipMessageNFS}
    Check Location Of Backup  ${EXTERNAL_BACKUP_PATH}  ${incnfsbackupjob}
    Check Duplicates Of Backups  /backup-storage/inc-backup-storage/  ${incnfsbackupjob}

Test Delete Incremental Data For Backup On External NFS Storage
    [Tags]  mongo  externalbackup
    Pass Execution If  not ${backup_mount}  ${skipMessageNFS}
    Pass Execution If  not ${full_restore}  ${skipMessage}
    Drop Mongodb Database  ${dbNameIncrementalNFS}
    Check ${documentIncrementalNFS} Should Not Exist In ${collectionNameIncrementalNFS} Of ${dbNameIncrementalNFS}

Test Restore Data From Incremental Backup On External NFS Storage
    [Tags]  mongo  externalbackup
    Pass Execution If  not ${backup_mount}  ${skipMessageNFS}
    Pass Execution If  not ${full_restore}  ${skipMessage}
    ${document}=  Set Variable  {"externalBackupPath":"${EXTERNAL_BACKUP_PATH}","vault":"${incnfsbackupjob}"}
    ${endpoint}=  Set Variable  /incremental
    Restore ${document} Data And Check With 100 Attempts On ${endpoint}

Test Check All Restored Incremental Data
    [Tags]  mongo  externalbackup
    Pass Execution If  not ${backup_mount}  ${skipMessageNFS}
    Pass Execution If  not ${full_restore}  ${skipMessage}
    Check ${documentIncrementalNFS} Should Exist In ${collectionNameIncrementalNFS} Of ${dbNameIncrementalNFS}

Test Delete Full Data For Backup Before Incremental On External NFS Storage
    [Tags]  mongo  externalbackup
    Pass Execution If  not ${backup_mount}  ${skipMessageNFS}
    Pass Execution If  not ${full_restore}  ${skipMessage}
    Drop Mongodb Database  ${dbNameNFS}
    Check ${documentNFS} Should Not Exist In ${collectionNameNFS} Of ${dbNameNFS}

Test Restore Data From Full Backup Before Incremental On External NFS Storage
    [Tags]  mongo  externalbackup
    Pass Execution If  not ${backup_mount}  ${skipMessageNFS}
    Pass Execution If  not ${full_restore}  ${skipMessage}
    ${document}=  Set Variable  {"externalBackupPath":"${EXTERNAL_BACKUP_PATH}","vault":"${fullnfsbackupjob}","dbs":["${dbNameNFS}"]}
    Restore ${document} Data And Check With 100 Attempts On ${Empty}

Test Check All Restored Full Data From Full Backup Before Incremental On External NFS Storage
    [Tags]  mongo  externalbackup
    Pass Execution If  not ${backup_mount}  ${skipMessageNFS}
    Pass Execution If  not ${full_restore}  ${skipMessage}
    Check ${documentNFS} Should Exist In ${collectionNameNFS} Of ${dbNameNFS}

