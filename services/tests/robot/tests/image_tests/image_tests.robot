*** Settings ***
Library    String
Library    Collections
Library  ../lib/KubernetesClient.py


*** Variables ***
${NAMESPACE}     %{OPENSHIFT_WORKSPACE_WA}
${CONFIG_NAME}                   %{CONFIG_NAME}
${SUPPLEMENTARY_CONFIG_NAME}     %{SUPPLEMENTARY_CONFIG_NAME}


*** Keywords ***
Compare Images From Resources With Dd
    [Arguments]  ${dd_images}
    ${stripped_resources}=  Strip String  ${dd_images}  characters=,  mode=right
    @{list_resources} =  Split String	${stripped_resources} 	,
    FOR  ${resource}  IN  @{list_resources}
      ${type}  ${name}  ${container_name}  ${image}=  Split String	${resource}
      ${resource_image}=  Get Resource Image  ${type}  ${name}  ${NAMESPACE}   ${container_name}
      Should Be Equal  ${resource_image}  ${image}
    END

*** Test Cases ***
Test Hardcoded Images
    [Tags]  check_mongo_images  smoke  mongo
    ${dd_images}=  Get Dd Images From Config Map  ${CONFIG_NAME}  ${NAMESPACE}
    Skip If  '${dd_images}' == '${None}'  There is no deployDescriptor, not possible to check case!
    Compare Images From Resources With Dd  ${dd_images}

Test Hardcoded Images For Supplementary Services
    [Tags]  check_mongo_images  smoke  mongo
    ${dd_images}=  Get Dd Images From Config Map  ${SUPPLEMENTARY_CONFIG_NAME}  ${NAMESPACE}
    Skip If  '${dd_images}' == '${None}'  There is no deployDescriptor, not possible to check case!
    Compare Images From Resources With Dd  ${dd_images}
