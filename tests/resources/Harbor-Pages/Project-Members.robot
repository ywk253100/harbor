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
Go Into Project
    [Arguments]  ${project}  ${has_image}=${true}
    Wait Until Element Is Visible And Enabled  ${search_input}
    Input Text  ${search_input}  ${project}
    Wait Until Page Contains  ${project}
    Wait Until Element Is Visible And Enabled  xpath=//*[@id='project-results']//clr-dg-cell[contains(.,'${project}')]/a
    Click Element  xpath=//*[@id='project-results']//clr-dg-cell[contains(.,'${project}')]/a
    #To prevent waiting for a fixed-period of time for page loading and failure caused by exception, we add loop to re-run <Wait Until Element Is Visible And Enabled> when
    #    exception was caught.
    :For  ${n}  IN RANGE  1  5
    \    ${out}  Run Keyword If  ${has_image}==${false}  Run Keyword And Ignore Error  Wait Until Element Is Visible And Enabled  xpath=//clr-dg-placeholder[contains(.,\"We couldn\'t find any repositories!\")]
    \    ...  ELSE  Run Keyword And Ignore Error  Wait Until Element Is Visible And Enabled  xpath=//clr-dg-cell[contains(.,'${project}/')]
    \    Log To Console  ${out[0]}
    \    ${result}  Set Variable If  '${out[0]}'=='PASS'  ${true}  ${false}
    \    Run Keyword If  ${result} == ${true}  Exit For Loop
    \    Sleep  1
    Should Be Equal  ${result}  ${true}

Add User To Project Admin
    [Arguments]  ${project}  ${user}
    Go Into Project
    Sleep  2  
    Click Element  xpath=${project_member_tag_xpath}
    Sleep  1
    Click Element  xpath=${project_member_add_button_xpath}
    Sleep  2	
    Input Text  xpath=${project_member_add_username_xpath}  ${user}
    Sleep  3
    Click Element  xpath=${project_member_add_admin_xpath}
    Click Element  xpath=${project_member_add_save_button_xpath}
    Sleep  4
    
Search Project Member
    [Arguments]  ${project}  ${user}
    Go Into Project  ${project}
    Sleep  2   
    Click Element  xpath=//clr-dg-cell//a[contains(.,'${project}')]
    Sleep  1	
    Click Element  xpath=${project_member_search_button_xpath}
    Sleep  1	
    Click Element  xpath=${project_member_search_text_xpath}
    Sleep  2
    Wait Until Page Contains  ${user}	
    
Change Project Member Role
    [Arguments]  ${project}  ${user}  ${role}
    Click Element  xpath=//clr-dg-cell//a[contains(.,'${project}')]
    Sleep  2    
    Click Element  xpath=${project_member_tag_xpath}
    Sleep  1	
    Click Element  xpath=//project-detail//clr-dg-row[contains(.,'${user}')]//label
    Sleep  1
    #change role
    Click Element  //*[@id='member-action']
    Click Element  //button[contains(.,'${role}')]
    Sleep  2
    Wait Until Page Contains  ${role}

User Can Change Role
     [arguments]  ${username}
     Click Element  xpath=//clr-dg-row[contains(.,'${username}')]//input/../label
     Click Element  xpath=//clr-dropdown[@id='member-action']
     Page Should Not Contain Element  xpath=//button[@disabled='' and contains(.,'Admin')]

User Can Not Change Role
     [arguments]  ${username}
     Click Element  xpath=//clr-dg-row[contains(.,'${username}')]//input/../label
     Click Element  xpath=//clr-dropdown[@id='member-action']
     Page Should Contain Element  xpath=//button[@disabled='' and contains(.,'Admin')]

#this keyworkd seems will not use any more, will delete in the future
Non-admin View Member Account
    [arguments]  ${times}
    Xpath Should Match X Times  //clr-dg-row-master  ${times}

User Can Not Add Member
    Page Should Contain Element  xpath=//button[@disabled='' and contains(.,'User')]

Add Guest Member To Project
    [arguments]  ${member}
    Click Element  xpath=${project_member_add_button_xpath}
    Sleep  1
    Input Text  xpath=${project_member_add_username_xpath}  ${member}
    #select guest
    Mouse Down  xpath=${project_member_guest_radio_checkbox}
    Mouse Up  xpath=${project_member_guest_radio_checkbox}
    Click Button  xpath=${project_member_add_confirmation_ok_xpath}
    Sleep  1

Delete Project Member
    [arguments]  ${member}
    Click Element  xpath=//clr-dg-row[contains(.,'${member}')]//input/../label
    Click Element  ${member_action_xpath}
    Sleep  1
    Click Element  ${delete_action_xpath}
    Sleep  2
    Click Element  //clr-modal//button[contains(.,'DELETE')]
    Sleep  3

User Should Be Owner Of Project
    [Arguments]  ${user}  ${pwd}  ${project}
    Sign In Harbor  ${HARBOR_URL}  ${user}  ${pwd}
    Go Into Project  ${project}
    Switch To Member
    User Can Not Change Role  ${user}
    Push image  ${ip}  ${user}  ${pwd}  ${project}  hello-world
    Logout Harbor

User Should Not Be A Member Of Project
    [Arguments]  ${user}  ${pwd}  ${project}
    Sign In Harbor  ${HARBOR_URL}  ${user}  ${pwd}
    Project Should Not Display  ${project}
    Logout Harbor
    Cannot Pull image  ${ip}  ${user}  ${pwd}  ${project}  ${ip}/${project}/hello-world
    Cannot Push image  ${ip}  ${user}  ${pwd}  ${project}  ${ip}/${project}/hello-world

Manage Project Member
    [Arguments]  ${admin}  ${pwd}  ${project}  ${user}  ${op}  ${has_image}=${true}
    Sign In Harbor  ${HARBOR_URL}  ${admin}  ${pwd}
    Go Into Project  ${project}  ${has_image}
    Switch To Member
    Run Keyword If  '${op}' == 'Add'  Add Guest Member To Project  ${user}
    ...    ELSE IF  '${op}' == 'Remove'  Delete Project Member  ${user}
    ...    ELSE  Change Project Member Role  ${project}  ${user}  ${role}
    Logout Harbor

Change User Role In Project
    [Arguments]  ${admin}  ${pwd}  ${project}  ${user}  ${role}
    Sign In Harbor  ${HARBOR_URL}  ${admin}  ${pwd}
    Wait Until Element Is Visible  //clr-dg-cell//a[contains(.,'${project}')]
    Change Project Member Role  ${project}  ${user}  ${role}
    Logout Harbor

User Should Be Guest
    [Arguments]  ${user}  ${pwd}  ${project}
    Sign In Harbor   ${HARBOR_URL}  ${user}  ${pwd}
    Project Should Display  ${project}
    Go Into Project  ${project}
    Switch To Member
    User Can Not Add Member
    Page Should Contain Element  xpath=//clr-dg-row[contains(.,'${user}')]//clr-dg-cell[contains(.,'Guest')]
    Logout Harbor
    Pull image  ${ip}  ${user}  ${pwd}  ${project}  hello-world
    Cannot Push image  ${ip}  ${user}  ${pwd}  ${project}  hello-world

User Should Be Developer
    [Arguments]  ${user}  ${pwd}  ${project}
    Sign In Harbor  ${HARBOR_URL}  ${user}  ${pwd}
    Project Should Display  ${project}
    Go Into Project  ${project}
    Switch To Member
    User Can Not Add Member
    Page Should Contain Element  xpath=//clr-dg-row[contains(.,'${user}')]//clr-dg-cell[contains(.,'Developer')]
    Logout Harbor
    Push Image With Tag  ${ip}  ${user}  ${pwd}  ${project}  hello-world  v1

User Should Be Admin
    [Arguments]  ${user}  ${pwd}  ${project}  ${guest}
    Sign In Harbor  ${HARBOR_URL}  ${user}  ${pwd}
    Project Should Display  ${project}
    Go Into Project  ${project}
    Switch To Member
    Add Guest Member To Project  ${guest}
    User Can Change Role  ${guest}
    Page Should Contain Element  xpath=//clr-dg-row[contains(.,'${user}')]//clr-dg-cell[contains(.,'Admin')]
    Logout Harbor
    Push Image With Tag  ${ip}  ${user}  ${pwd}  ${project}  hello-world  v2

Project Should Have Member
    [Arguments]  ${project}  ${user}
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
    Go Into Project  ${project}
    Switch To Member
    Page Should Contain Element  xpath=//clr-dg-cell[contains(., '${user}')]
    Logout Harbor
