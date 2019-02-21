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
Library  OperatingSystem
Library  String
Library  Collections
Library  requests
Library  Process
Library  SSHLibrary  1 minute
Library  DateTime
Library  Selenium2Library  60  10
Library  JSONLibrary
Resource  Nimbus-Util.robot
Resource  Vsphere-Util.robot
Resource  VCH-Util.robot
Resource  Drone-Util.robot
Resource  Github-Util.robot
Resource  Harbor-Util.robot
Resource  Harbor-Pages/HomePage.robot
Resource  Harbor-Pages/HomePage_Elements.robot
Resource  Harbor-Pages/Project.robot
Resource  Harbor-Pages/Project_Elements.robot
Resource  Harbor-Pages/Project-Members.robot
Resource  Harbor-Pages/Project-Members_Elements.robot
Resource  Harbor-Pages/Project-Repository.robot
Resource  Harbor-Pages/Project-Repository_Elements.robot
Resource  Harbor-Pages/Project-Config.robot
Resource  Harbor-Pages/Project-Helmcharts.robot
Resource  Harbor-Pages/Project-Helmcharts_Elements.robot
Resource  Harbor-Pages/Project-Retag.robot
Resource  Harbor-Pages/Project-Retag_Elements.robot
Resource  Harbor-Pages/Replication.robot
Resource  Harbor-Pages/Replication_Elements.robot
Resource  Harbor-Pages/UserProfile.robot
Resource  Harbor-Pages/Administration-Users.robot
Resource  Harbor-Pages/Administration-Users_Elements.robot
Resource  Harbor-Pages/Configuration.robot
Resource  Harbor-Pages/Configuration_Elements.robot
Resource  Harbor-Pages/ToolKit.robot
Resource  Harbor-Pages/ToolKit_Elements.robot
Resource  Harbor-Pages/Vulnerability.robot
Resource  Harbor-Pages/LDAP-Mode.robot
Resource  Harbor-Pages/Verify.robot
Resource  Docker-Util.robot
Resource  Admiral-Util.robot
Resource  OVA-Util.robot
Resource  Cert-Util.robot
Resource  SeleniumUtil.robot
Resource  Nightly-Util.robot
Resource  APITest-Util.robot

*** Keywords ***
Wait Until Element Is Visible And Enabled
    [Arguments]  ${element}
    Wait Until Element Is Visible  ${element}
    Wait Until Element Is Enabled  ${element}

Retry Action Keyword
    [Arguments]  ${keyword}  ${element_xpath}
    Retry Keyword When Error  ${keyword}  element=${element_xpath}

Retry Element Click
    [Arguments]  ${element_xpath}
    Retry Action Keyword  Element Click  ${element_xpath}

Retry Wait Element
    [Arguments]  ${element_xpath}
    Retry Action Keyword  Wait Until Element Is Visible And Enabled  ${element_xpath}

Retry Wait Element Not Visible
    [Arguments]  ${element_xpath}
    Retry Action Keyword  Wait Until Element Is Not Visible  ${element_xpath}

Retry Button Click
    [Arguments]  ${element_xpath}
    Retry Action Keyword  Button Click  ${element_xpath}

Element Click
    [Arguments]  ${element_xpath}
    Wait Until Element Is Visible And Enabled  ${element_xpath}
    Click Element  ${element_xpath}

Button Click
    [Arguments]  ${element_xpath}
    Wait Until Element Is Visible And Enabled  ${element_xpath}
    Click button  ${element_xpath}

Wait Unitl Vul Data Ready
    [Arguments]  ${url}  ${timeout}  ${interval}
    ${n}=  Evaluate  ${timeout}/${interval}
    :FOR  ${i}  IN RANGE  ${n}
    \    Log  Checking the vul data: ${i} ...  console=True
    \    ${rc}  ${output}=  Run And Return Rc And Output  curl -k ${url}/api/systeminfo
    \    Should Be Equal As Integers  ${rc}  0
    \    ${contains}=  Run Keyword And Return Status  Should Contain  ${output}  overall_last_update
    \    Exit For Loop If  ${contains}
    \    Sleep  ${interval}
    Run Keyword If  ${i+1}==${n}  Fail  The vul data is not ready

Wait Unitl Command Success
    [Arguments]  ${cmd}  ${times}=8  ${positive}=${true}
    :FOR  ${n}  IN RANGE  1  ${times}
    \    Log  Trying ${cmd}: ${n} ...  console=True
    \    ${rc}  ${output}=  Run And Return Rc And Output  ${cmd}
    \    Run Keyword If  ${positive} == ${true}  Exit For Loop If  '${rc}'=='0'
    \    ...  ELSE  Exit For Loop If  '${rc}'!='0'
    \    Sleep  2
    Log  ${output}
    Run Keyword If  ${positive} == ${true}  Should Be Equal As Strings  '${rc}'  '0'
    ...  ELSE  Should Not Be Equal As Strings  '${rc}'  '0'   
    [Return]  ${output}

Retry Keyword When Error
    [Arguments]  ${keyword}  ${element}=${None}  ${times}=6
    :For  ${n}  IN RANGE  1  ${times}
    \    Log To Console  Trying ${keyword} ${n} times ...
    \    ${out}  Run Keyword If  "${element}"=="${None}"  Run Keyword And Ignore Error  ${keyword}
    \    ...  ELSE  Run Keyword And Ignore Error  ${keyword}  ${element}
    \    Log To Console  Return value is ${out[0]}
    \    Exit For Loop If  '${out[0]}'=='PASS'
    \    Sleep  2
    Run Keyword If  '${out[0]}'=='FAIL'  Capture Page Screenshot
    Should Be Equal As Strings  '${out[0]}'  'PASS'