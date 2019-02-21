# Copyright Project Harbor Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#	http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License

*** Settings ***
Documentation  This resource provides any keywords related to the Harbor private registry appliance
Resource  ../../resources/Util.robot

*** Variables ***
${HARBOR_VERSION}  v1.1.1

*** Keywords ***
Init LDAP
    ${rc}  ${output}=  Run And Return Rc And Output  ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'
    Log  ${output}
    Sleep  2
    Input Text  xpath=//*[@id='ldapUrl']  ldaps://${output}
    Sleep  1
    Input Text  xpath=//*[@id='ldapSearchDN']  cn=admin,dc=example,dc=com
    Sleep  1
    Input Text  xpath=//*[@id='ldapSearchPwd']  admin
    Sleep  1
    Input Text  xpath=//*[@id='ldapBaseDN']  dc=example,dc=com
    Sleep  1
    Input Text  xpath=//*[@id='ldapFilter']  (&(objectclass=inetorgperson)(memberof=cn=harbor_users,ou=groups,dc=example,dc=com))
    Sleep  1
    Input Text  xpath=//*[@id='ldapUid']  cn
    Sleep  1
    Capture Page Screenshot
    Disable Ldap Verify Cert Checkbox
    Click Element  xpath=${config_auth_save_button_xpath}
    Sleep  2
    Click Element  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/div/div/config/div/div/div/button[3]
    Sleep  1
    Capture Page Screenshot

Switch To Configure
    Click Element  xpath=${configuration_xpath}
    Sleep  2

Test Ldap Connection
    ${rc}  ${output}=  Run And Return Rc And Output  ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'
    Log  ${output}
    Sleep  2
    Input Text  xpath=//*[@id='ldapUrl']  ldaps://${output}
    Sleep  1
    Input Text  xpath=//*[@id='ldapSearchDN']  cn=admin,dc=example,dc=com
    Sleep  1
    Input Text  xpath=//*[@id='ldapSearchPwd']  admin
    Sleep  1
    Input Text  xpath=//*[@id='ldapBaseDN']  dc=example,dc=com
    Sleep  1
    Input Text  xpath=//*[@id='ldapUid']  cn
    Sleep  1

    # default is checked, click test connection to verify fail as no cert.
    Click Element  xpath=${test_ldap_xpath}
    Sleep  1
    Wait Until Page Contains  Failed to verify LDAP server with error
    Sleep  5

    Disable Ldap Verify Cert Checkbox
    # ldap checkbox unchecked, click test connection to verify success.
    Sleep  1
    Click Element  xpath=${test_ldap_xpath}
    Capture Page Screenshot
    Wait Until Page Contains  Connection to LDAP server is verified  timeout=15

Test LDAP Server Success
    Click Element  xpath=${test_ldap_xpath}
    Wait Until Page Contains  Connection to LDAP server is verified  timeout=15

Disable Ldap Verify Cert Checkbox
    Mouse Down  xpath=//*[@id='clr-checkbox-ldapVerifyCert']
    Mouse Up  xpath=//*[@id='clr-checkbox-ldapVerifyCert']
    Sleep  2
    Capture Page Screenshot
    Ldap Verify Cert Checkbox Should Be Disabled

Ldap Verify Cert Checkbox Should Be Disabled
    Checkbox Should Not Be Selected  xpath=//*[@id='clr-checkbox-ldapVerifyCert']

Set Pro Create Admin Only	
    #set limit to admin only
    Click Element  xpath=${configuration_xpath}
    Sleep  2
    Click Element  xpath=${system_config_xpath}
    Sleep  1
    Click Element  xpath=//select[@id='proCreation']
    Click Element  xpath=//select[@id='proCreation']//option[@value='adminonly']
    Sleep  1
    Click Element  xpath=${config_system_save_button_xpath}
    Capture Page Screenshot  AdminCreateOnly.png

Set Pro Create Every One
    Click Element  xpath=${configuration_xpath}
    sleep  1
    #set limit to Every One
    Click Element  xpath=${system_config_xpath}
    Sleep  1
    Click Element  xpath=//select[@id='proCreation']
    Click Element  xpath=//select[@id='proCreation']//option[@value='everyone']
    Sleep  1	
    Click Element  xpath=${config_system_save_button_xpath}
    Sleep  2
    Capture Page Screenshot  EveryoneCreate.png

Disable Self Reg
    Click Element  xpath=${configuration_xpath}
    Mouse Down  xpath=${self_reg_xpath}
    Mouse Up  xpath=${self_reg_xpath}
    Sleep  1
    Self Reg Should Be Disabled
    Click Element  xpath=${config_auth_save_button_xpath}
    Capture Page Screenshot  DisableSelfReg.png
    Sleep  1

Enable Self Reg	
    Mouse Down  xpath=${self_reg_xpath}
    Mouse Up  xpath=${self_reg_xpath}
    Sleep  1
    Self Reg Should Be Enabled
    Click Element  xpath=${config_auth_save_button_xpath}
    Capture Page Screenshot  EnableSelfReg.png
    Sleep  1

Self Reg Should Be Disabled
    Checkbox Should Not Be Selected  xpath=${self_reg_xpath}

Self Reg Should Be Enabled
    Checkbox Should Be Selected  xpath=${self_reg_xpath}

Project Creation Should Display
    Page Should Contain Element  xpath=${project_create_xpath}

Project Creation Should Not Display
    Page Should Not Contain Element  xpath=${project_create_xpath}

## System settings	
Switch To System Settings
    Sleep  1
    Click Element  xpath=${configuration_xpath}
    Click Element  xpath=${system_config_xpath}
Modify Token Expiration
    [Arguments]  ${minutes}
    Input Text  xpath=//*[@id='tokenExpiration']  ${minutes}
    Click Button  xpath=${config_system_save_button_xpath} 
    Sleep  1

Token Must Be Match
    [Arguments]  ${minutes}
    Textfield Value Should Be  xpath=//*[@id='tokenExpiration']  ${minutes}

## Replication	
Check Verify Remote Cert	
    Mouse Down  xpath=//*[@id='clr-checkbox-verifyRemoteCert'] 
    Mouse Up  xpath=//*[@id='clr-checkbox-verifyRemoteCert']
    Click Element  xpath=${config_save_button_xpath}
    Capture Page Screenshot  RemoteCert.png
    Sleep  1

Switch To System Replication
    Sleep  1
    Switch To Configure
    Click Element  xpath=//*[@id='config-replication']
    Sleep  1

Should Verify Remote Cert Be Enabled
    Checkbox Should Not Be Selected  xpath=//*[@id='clr-checkbox-verifyRemoteCert']

## Email	
Switch To Email
    Switch To Configure
    Click Element  xpath=//*[@id='config-email']
    Sleep  1

Config Email
    Input Text  xpath=//*[@id='mailServer']  smtp.vmware.com
    Input Text  xpath=//*[@id='emailPort']  25
    Input Text  xpath=//*[@id='emailUsername']  example@vmware.com 
    Input Text  xpath=//*[@id='emailPassword']  example
    Input Text  xpath=//*[@id='emailFrom']  example<example@vmware.com>
    Sleep  1    
    Click Element  xpath=//clr-checkbox-wrapper[@id='emailSSL-wrapper']//label
    Sleep  1
    Click Element  xpath=//clr-checkbox-wrapper[@id='emailInsecure-wrapper']//label
    Sleep  1
    Click Element  xpath=${config_email_save_button_xpath}
    Sleep  6

Verify Email
    Textfield Value Should Be  xpath=//*[@id='mailServer']  smtp.vmware.com
    Textfield Value Should Be  xpath=//*[@id='emailPort']  25
    Textfield Value Should Be  xpath=//*[@id='emailUsername']  example@vmware.com
    Textfield Value Should Be  xpath=//*[@id='emailFrom']  example<example@vmware.com>
    Checkbox Should Be Selected  xpath=//*[@id='emailSSL']
    Checkbox Should Not Be Selected  xpath=//*[@id='emailInsecure']

Set Scan All To None
    click element  //vulnerability-config//select
    click element  //vulnerability-config//select/option[@value='none']
    sleep  1
    click element  ${vulnerbility_save_button_xpath}

Set Scan All To Daily
    click element  //vulnerability-config//select
    click element  //vulnerability-config//select/option[@value='daily']
    sleep  1
    click element  ${vulnerbility_save_button_xpath}

Click Scan Now
    click element  //vulnerability-config//button[contains(.,'SCAN')]


Enable Read Only
    ${rc}  ${output}=  Run And Return Rc And Output  curl -u admin:Harbor12345 -s --insecure -H "Content-Type: application/json" -X PUT -d '{"read_only":true}' "https://${ip}/api/configurations"
    Log To Console  ${output}
    Should Be Equal As Integers  ${rc}  0

Disable Read Only
    ${rc}  ${output}=  Run And Return Rc And Output  curl -u admin:Harbor12345 -s --insecure -H "Content-Type: application/json" -X PUT -d '{"read_only":false}' "https://${ip}/api/configurations"
    Log To Console  ${output}
    Should Be Equal As Integers  ${rc}  0

## System labels
Switch To System Labels
    Sleep  1
    Click Element  xpath=${configuration_xpath}
    Click Element  xpath=//*[@id='config-label']

Create New Labels
    [Arguments]  ${labelname}
    Click Element  xpath=//button[contains(.,'New Label')]
    Sleep  1
    Input Text  xpath=//*[@id='name']  ${labelname}
    Sleep  1
    Click Element  xpath=//hbr-create-edit-label//clr-dropdown/clr-icon
    Sleep  1
    Click Element  xpath=//hbr-create-edit-label//clr-dropdown-menu/label[1]
    Sleep  1
    Input Text  xpath=//*[@id='description']  global
    Click Element  xpath=//div/form/section/label[4]/button[2]
    Capture Page Screenshot
    Wait Until Page Contains  ${labelname}

Update A Label
    [Arguments]  ${labelname}
    Click Element  xpath=//clr-dg-row[contains(.,'${labelname}')]//clr-checkbox-wrapper
    Sleep  1
    Click Element  xpath=//button[contains(.,'Edit')]
    Sleep  1
    Input Text  xpath=//*[@id='name']  ${labelname}1
    Sleep  1
    Click Element  xpath=//hbr-create-edit-label//form/section//button[2]
    Capture Page Screenshot
    Wait Until Page Contains  ${labelname}1

Delete A Label
    [Arguments]  ${labelname}
    Click Element  xpath=//clr-dg-row[contains(.,'${labelname}')]//clr-checkbox-wrapper
    Sleep  1
    Click ELement  xpath=//button[contains(.,'Delete')]
    Sleep  3
    Capture Page Screenshot
    Click Element  xpath=//clr-modal//div//button[contains(.,'DELETE')]
    Wait Until Page Contains Element  //clr-tab-content//div[contains(.,'${labelname}')]/../div/clr-icon[@shape='success-standard']

## Garbage Collection	
Switch To Garbage Collection
    Sleep  1
    Click Element  xpath=${configuration_xpath}
    Wait Until Page Contains Element  ${garbage_collection_xpath} 
    Click Element  xpath=${garbage_collection_xpath}

Click GC Now
    Sleep  1
    Click Element  xpath=${gc_now_xpath}
    Sleep  2

View GC Details
    Click Element  xpath=${gc_log_details_xpath}
    Sleep  2
