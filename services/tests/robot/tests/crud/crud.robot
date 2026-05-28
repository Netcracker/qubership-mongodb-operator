*** Variables ***

*** Settings ***
Library  String
Library    OperatingSystem
Resource  ../shared/keywords.robot
Suite Setup  Preparation
Suite Teardown  Cleanup

*** Keywords ***
Preparation
    Prepare Shared

    ${collection}=  Generate Random String  10  [LOWER]
    Set Suite Variable  ${collection}

    ${docValueOne}=  Generate Random String  10  [LOWER]
    ${docValueTwo}=  Generate Random String  10  [LOWER]
    ${docUpdatedValueOne}=  Generate Random String  10  [LOWER]

    ${documentOne}=  Set Variable  {"name":"${docValueOne}"}
    ${documentTwo}=  Set Variable  {"name":"${docValueTwo}"}
    ${documentUpdatedOne}=  Set Variable  {"name":"${docUpdatedValueOne}"}
    Set Suite Variable  ${documentOne}
    Set Suite Variable  ${documentTwo}
    Set Suite Variable  ${documentUpdatedOne}


Cleanup
    Teardown Shared

*** Test Cases ***
Test Insert Documents
    [Tags]  mongo  smoke
    Insert One Mongodb Document  ${collection}  ${documentOne}  database_name=${MONGO_DB}
    Insert One Mongodb Document  ${collection}  ${documentTwo}  database_name=${MONGO_DB}

    ${collectionsList}=  Get Collection Names  database_name=${MONGO_DB}
    Log  ${collectionsList}
    Should Contain  b"${collectionsList}"  ${collection}

    Check ${documentOne} Should Exist In ${collection} Of ${MONGO_DB}

    Check ${documentTwo} Should Exist In ${collection} Of ${MONGO_DB}

Test Update Document
    [Tags]  mongo  smoke
    Update One Mongodb Document  ${collection}  ${documentOne}  ${documentUpdatedOne}  set  database_name=${MONGO_DB}

    Check ${documentOne} Should Not Exist In ${collection} Of ${MONGO_DB}
    Check ${documentUpdatedOne} Should Exist In ${collection} Of ${MONGO_DB}

Test Delete Document
    [Tags]  mongo  smoke
    Delete One Mongodb Document  ${collection}  ${documentTwo}  database_name=${MONGO_DB}

    Check ${documentTwo} Should Not Exist In ${collection} Of ${MONGO_DB}

Test Drop Collection
    [Tags]  mongo  smoke
    Drop Mongodb Collection  ${collection}  database_name=${MONGO_DB}

    Check ${documentUpdatedOne} Should Not Exist In ${collection} Of ${MONGO_DB}