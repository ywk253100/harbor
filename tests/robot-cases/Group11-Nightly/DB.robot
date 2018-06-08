// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

*** Settings ***
Documentation  Harbor BATs
Resource  ../../resources/Util.robot
Default Tags  Nightly

*** Variables ***
${HARBOR_URL}  https://${ip}
${SSH_USER}  root
${HARBOR_ADMIN}  admin

*** Test Cases ***
Test Case - Create An New User
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Create An New User  url=${HARBOR_URL}  username=tester${d}  email=tester${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest
    Close Browser

Test Case - User View Projects
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Create An New User  url=${HARBOR_URL}  username=tester${d}  email=tester${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest
    Create An New Project  test${d}1
    Create An New Project  test${d}2
    Create An New Project  test${d}3
    Switch To Log
    Wait Until Page Contains  test${d}1
    Wait Until Page Contains  test${d}2
    Wait Until Page Contains  test${d}3
    Close Browser

Test Case - Update User Comment
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Create An New User  url=${HARBOR_URL}  username=tester${d}  email=tester${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest
    Update User Comment  Test12#4
    Logout Harbor

Test Case - Update Password
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Create An New User  url=${HARBOR_URL}  username=tester${d}  email=tester${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest
    Change Password  Test1@34  Test12#4
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  tester${d}  Test12#4
    Close Browser

Test Case - User View Logs
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s

    Create An New Project With New User  url=${HARBOR_URL}  username=tester${d}  email=tester${d}@vmware.com  realname=tester${d}  newPassword=Test1@34  comment=harbor  projectname=project${d}  public=true

    Push image  ${ip}  tester${d}  Test1@34  project${d}  busybox:latest
    Pull image  ${ip}  tester${d}  Test1@34  project${d}  busybox:latest

    Go Into Project  project${d}
    Delete Repo  project${d}

    Go To Project Log
    Advanced Search Should Display

    Do Log Advanced Search
    Close Browser

Test Case - Delete Multi User
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s
    Create An New User  ${HARBOR_URL}  deletea${d}  testa${d}@vmware.com  test${d}  Test1@34  harbor
    Logout Harbor
    Create An New User  ${HARBOR_URL}  deleteb${d}  testb${d}@vmware.com  test${d}  Test1@34  harbor
    Logout Harbor
    Create An New User  ${HARBOR_URL}  deletec${d}  testc${d}@vmware.com  test${d}  Test1@34  harbor
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  admin  Harbor12345
    Switch To User Tag
    Filter Object  delete
    Multi-delete Object  deletea  deleteb  deletec
    # Assert delete
    Delete Success  deletea  deleteb  deletec
    Sleep  1
    Close Browser
    
Test Case - Edit Self-Registration
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Disable Self Reg
    Logout Harbor

    Sign Up Should Not Display

    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Configure
    Self Reg Should Be Disabled
    Sleep  1

    # Restore setting
    Enable Self Reg
    Close Browser

Test Case - Manage Project Member
    Init Chrome Driver
    ${d}=    Get current Date  result_format=%m%s

    Create An New Project With New User  url=${HARBOR_URL}  username=alice${d}  email=alice${d}@vmware.com  realname=alice${d}  newPassword=Test1@34  comment=harbor  projectname=project${d}  public=false
    Push image  ip=${ip}  user=alice${d}  pwd=Test1@34  project=project${d}  image=hello-world
    Logout Harbor
    Create An New User  url=${HARBOR_URL}  username=bob${d}  email=bob${d}@vmware.com  realname=bob${d}  newPassword=Test1@34  comment=habor
    Logout Harbor
    Create An New User  url=${HARBOR_URL}  username=carol${d}  email=carol${d}@vmware.com  realname=carol${d}  newPassword=Test1@34  comment=harbor
    Logout Harbor

    User Should Not Be A Member Of Project  bob${d}  Test1@34  project${d}
    Manage Project Member  alice${d}  Test1@34  project${d}  bob${d}  Add
    User Should Be Guest  bob${d}  Test1@34  project${d}
    Change User Role In Project  alice${d}  Test1@34  project${d}  bob${d}  Developer
    User Should Be Developer  bob${d}  Test1@34  project${d}
    Change User Role In Project  alice${d}  Test1@34  project${d}  bob${d}  Admin
    User Should Be Admin  bob${d}  Test1@34  project${d}  carol${d}
    Manage Project Member  alice${d}  Test1@34  project${d}  bob${d}  Remove
    User Should Not Be A Member Of Project  bob${d}  Test1@34  project${d}
    User Should Be Guest  carol${d}  Test1@34  project${d}

    Close Browser

Test Case - Manage project publicity
    Init Chrome Driver
    ${d}=    Get Current Date  result_format=%m%s

    Create An New User  url=${HARBOR_URL}  username=usera${d}  email=usera${d}@vmware.com  realname=usera${d}  newPassword=Test1@34  comment=harbor
    Logout Harbor
    Create An New User  url=${HARBOR_URL}  username=userb${d}  email=userb${d}@vmware.com  realname=userb${d}  newPassword=Test1@34  comment=harbor
    Logout Harbor

    Sign In Harbor  ${HARBOR_URL}  usera${d}  Test1@34
    Create An New Project  project${d}  public=true

    Push image  ${ip}  usera${d}  Test1@34  project${d}  hello-world:latest
    Pull image  ${ip}  userb${d}  Test1@34  project${d}  hello-world:latest

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  userb${d}  Test1@34
    Project Should Display  project${d}
    Search Private Projects
    Project Should Not Display  project${d}

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  usera${d}  Test1@34
    Make Project Private  project${d}

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  userb${d}  Test1@34
    Project Should Not Display  project${d}
    Cannot Pull image  ${ip}  userb${d}  Test1@34  project${d}  hello-world:latest

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  usera${d}  Test1@34
    Make Project Public  project${d}

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  userb${d}  Test1@34
    Project Should Display  project${d}
    Close Browser

Test Case - Assign Sys Admin
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Create An New User  url=${HARBOR_URL}  username=tester${d}  email=tester${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch to User Tag
    Assign User Admin  tester${d}
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  tester${d}  Test1@34
    Administration Tag Should Display
    Close Browser