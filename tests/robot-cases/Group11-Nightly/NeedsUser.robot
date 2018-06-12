### need users ...
Test Case - User View Projects
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  user001  Test1@34
    Create An New Project  test${d}1
    Create An New Project  test${d}2
    Create An New Project  test${d}3
    Switch To Log
    Wait Until Page Contains  test${d}1
    Wait Until Page Contains  test${d}2
    Wait Until Page Contains  test${d}3
    Close Browser

Test Case - User View Logs
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user002  Test1@34
    Create An New Project  project${d}
    
    Push image  ${ip}  tester${d}  Test1@34  project${d}  busybox:latest
    Pull image  ${ip}  tester${d}  Test1@34  project${d}  busybox:latest

    Go Into Project  project${d}
    Delete Repo  project${d}

    Go To Project Log
    Advanced Search Should Display

    Do Log Advanced Search
    Close Browser

Test Case - Manage Project Member
    Init Chrome Driver
    ${d}=    Get current Date  result_format=%m%s
 
    Sign In Harbor  ${HARBOR_URL}  user003  Test1@34
    Create An New Project  project${d}
    Push image  ip=${ip}  user=user004  pwd=Test1@34  project=project${d}  image=hello-world
    Logout Harbor

    User Should Not Be A Member Of Project  user005  Test1@34  project${d}
    Manage Project Member  alice${d}  Test1@34  project${d}  user005  Add
    User Should Be Guest  user005  Test1@34  project${d}
    Change User Role In Project  alice${d}  Test1@34  project${d}  user005  Developer
    User Should Be Developer  user005  Test1@34  project${d}
    Change User Role In Project  alice${d}  Test1@34  project${d}  user005  Admin
    User Should Be Admin  user005  Test1@34  project${d}  user006
    Manage Project Member  alice${d}  Test1@34  project${d}  user005  Remove
    User Should Not Be A Member Of Project  user005  Test1@34  project${d}
    User Should Be Guest  user006  Test1@34  project${d}

    Close Browser

Test Case - Manage project publicity
    Init Chrome Driver
    ${d}=    Get Current Date  result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user007  Test1@34
    Create An New Project  project${d}  public=true

    Push image  ${ip}  user007  Test1@34  project${d}  hello-world:latest
    Pull image  ${ip}  user008  Test1@34  project${d}  hello-world:latest

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  user008  Test1@34
    Project Should Display  project${d}
    Search Private Projects
    Project Should Not Display  project${d}

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  user007  Test1@34
    Make Project Private  project${d}

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  user008  Test1@34
    Project Should Not Display  project${d}
    Cannot Pull image  ${ip}  user008  Test1@34  project${d}  hello-world:latest

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  user007  Test1@34
    Make Project Public  project${d}

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  user008  Test1@34
    Project Should Display  project${d}
    Close Browser

Test Case - Assign Sys Admin
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch to User Tag
    Assign User Admin  user009
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  user009  Test1@34
    Administration Tag Should Display
    Close Browser

Test Case - Edit Project Creation
    # Create normal user and login
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  user010  Test1@34
    Project Creation Should Display
    Logout Harbor

    Sleep  3
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Set Pro Create Admin Only
    Logout Harbor

    Sign In Harbor  ${HARBOR_URL}  user010  Test1@34
    Project Creation Should Not Display
    Logout Harbor

    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Set Pro Create Every One
    Close browser

Test Case - Edit Repo Info
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    
    Sign In Harbor  ${HARBOR_URL}  user011  Test1@34
    Create An New Project  projectb${d}
    Push Image  ${ip}  user011  Test1@34  project${d}  hello-world
    Go Into Project  project${d}
    Go Into Repo  project${d}/hello-world
    Edit Repo Info
    Close Browser

Test Case - Delete Multi Project
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    
    Sign In Harbor  ${HARBOR_URL}  user012  Test1@34
    Create An New Project  projecta${d}
    Create An New Project  projectb${d}
    Push Image  ${ip}  test${d}  Test1@34  projecta${d}  hello-world
    Filter Object  project
    Multi-delete Object  projecta  projectb
    # Verify delete project with image should not be deleted directly
    Delete Fail  projecta${d}
    Delete Success  projectb${d}
    Close Browser

Test Case - Delete Multi Repo
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s
    
    Sign In Harbor  ${HARBOR_URL}  user013  Test1@34
    Create An New Project  project${d}
    Push Image  ${ip}  user013  Test1@34  project${d}  hello-world
    Push Image  ${ip}  user013  Test1@34  project${d}  busybox
    Sleep  2
    Go Into Project  project${d}
    Multi-delete Object  hello-world  busybox
    # Verify
    Delete Success  hello-world  busybox
    Close Browser

Test Case - Delete Multi Tag
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s
    
    Sign In Harbor  ${HARBOR_URL}  user014  Test1@34
    Create An New Project  project${d}
    Push Image With Tag  ${ip}  user014  Test1@34  project${d}  redis  3.2.10-alpine  3.2.10-alpine
    Push Image With Tag  ${ip}  user014  Test1@34  project${d}  redis  4.0.7-alpine  4.0.7-alpine
    Sleep  2
    Go Into Project  project${d}
    Go Into Repo  redis
    Multi-delete object  3.2.10-alpine  4.0.7-alpine
    # Verify
    Delete Success  3.2.10-alpine  4.0.7-alpine
    Close Browser

Test Case - Delete Repo on CardView
    Init Chrome Driver
    ${d}=   Get Current Date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  user015  Test1@34
    Create An New Project  project${d}
    Push Image  ${ip}  test${d}  Test1@34  project${d}  hello-world
    Push Image  ${ip}  test${d}  Test1@34  project${d}  busybox
    Sleep  2
    Go Into Project  project${d}
    Switch To CardView
    Delete Repo on CardView  busybox
    # Verify
    Delete Success  busybox
    Close Browser

Test Case - Delete Multi Member
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  user016  Test1@34    
    Create An New Project  project${d}
    Go Into Project  project${d}
    Switch To Member
    Add Guest Member to project  user017
    Add Guest Member to project  user018
    Multi-delete Member  user017  user018
    Delete Success  user017  user018
    Close Browser

Test Case - Project Admin Operate Labels
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  user019  Test1@34
    Create An New Project  project${d}
    Go Into Project  project${d}
    Sleep  2
    # Add labels
    Switch To Project Label
    Create New Labels  label_${d}
    Sleep  2
    Update A Label  label_${d}
    Sleep  2
    Delete A Label  label_${d}
    Close Browser

Test Case - Project Admin Add Labels To Repo
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  user020  Test1@34
    Create An New Project  project${d}
    Push Image With Tag  ${ip}  user020  Test1@34  project${d}  redis  3.2.10-alpine  3.2.10-alpine
    Push Image With Tag  ${ip}  user020  Test1@34  project${d}  redis  4.0.7-alpine  4.0.7-alpine

    Go Into Project  project${d}
    Sleep  2
    # Add labels
    Switch To Project Label
    Create New Labels  label111
    Create New Labels  label22
    Sleep  2
    Switch To Project Repo
    Go Into Repo  project${d}/redis
    Add Labels To Tag  3.2.10-alpine  label111
    Add Labels To Tag  4.0.7-alpine  label22
    Filter Labels In Tags  label111  label22
    Close Browser

Test Case - Developer Operate Labels
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user022  Test1@34
    Manage Project Member  user021  Test1@34  project${d}  user022  Add
    Change User Role In Project  user021  Test1@34  project${d}  user022  Developer

    Sign In Harbor  ${HARBOR_URL}  user022  Test1@34
    Go Into Project  project${d}
    Sleep  3
    Page Should Not Contain Element  xpath=//a[contains(.,'Labels')]
    Close Browser

Test Case - Scan A Tag In The Repo
    Init Chrome Driver
    ${d}=  get current date  result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user023  Test1@34
    Go Into Project  project${d}    
    Push Image  ${ip}  user023  Test1@34  project${d}  hello-world
    Go Into Project  project${d}
    Go Into Repo  project${d}/hello-world
    Scan Repo  latest  Succeed
    Summary Chart Should Display  latest
    Pull Image  ${ip}  user023  Test1@34  project${d}  hello-world
    # Edit Repo Info
    Close Browser

Test Case - Scan As An Unprivileged User
    Init Chrome Driver
    ${d}=    get current date    result_format=%m%s
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  library  hello-world
 
    Sign In Harbor  ${HARBOR_URL}  user024  Test1@34
    Go Into Project  library
    Go Into Repo  hello-world
    Select Object  latest
    Scan Is Disabled
    Close Browser

Test Case - Scan Image With Empty Vul
    Init Chrome Driver
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  library  hello-world
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Go Into Project  library
    Go Into Repo  hello-world
    Scan Repo  latest  Succeed
    Move To Summary Chart
    Wait Until Page Contains  Unknow
    Close Browser

Test Case - Manual Scan All
    Init Chrome Driver
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  library  redis
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Configure
    Go To Vulnerability Config
    Trigger Scan Now
    Back To Projects
    Go Into Project  library
    Go Into Repo  redis
    Summary Chart Should Display  latest
    Close Browser

Test Case - Scan Image On Push
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Go Into Project  library
    Goto Project Config
    Enable Scan On Push
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  library  memcached
    Back To Projects
    Go Into Project  library
    Go Into Repo  memcached
    Summary Chart Should Display  latest
    Close Browser

Test Case - View Scan Results
    Init Chrome Driver
    ${d}=  get current date  result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user025  Test1@34
    Create An New Project  project${d}    
    Push Image  ${ip}  user025  Test1@34  project${d}  tomcat
    Go Into Project  project${d}
    Go Into Repo  project${d}/tomcat
    Scan Repo  latest  Succeed
    Summary Chart Should Display  latest
    View Repo Scan Details
    Close Browser

Test Case - View Scan Error
    Init Chrome Driver
    ${d}=  get current date  result_format=%m%s
    Create An New Project With New User  url=${HARBOR_URL}  username=tester${d}  email=tester${d}@vmware.com  realname=tester${d}  newPassword=Test1@34  comment=harbor  projectname=project${d}  public=false

    Sign In Harbor  ${HARBOR_URL}  user026  Test1@34
    Create An New Project  project${d}   
    Push Image  ${ip}  tester${d}  Test1@34  project${d}  vmware/photon:1.0
    Go Into Project  project${d}
    Go Into Repo  project${d}/vmware/photon
    Scan Repo  1.0  Fail
    View Scan Error Log
    Close Browser