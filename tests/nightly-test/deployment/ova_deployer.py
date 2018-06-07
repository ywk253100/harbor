import logging
import os
import sys
import time
import subprocess
from datetime import datetime
dir_path = os.path.dirname(os.path.realpath(__file__))
sys.path.append(dir_path + '../utils')
import govc_utils
import nlogging
import urllib
import socket
import harbor_util
import buildweb_utils
logger = nlogging.create_logger(__name__)
import json

class OVADeployer():

    def __init__(self, auth_mode, ova_path, count=1):

        if ova_path == "latest" :
            self.ova_path = buildweb_utils.get_latest_build_url('master','beta')

        with open(os.getcwd() + '/tests/nightly-test/configuration/ova.json') as ova_config:
            ova_data = json.load(ova_config)
        
        self.vc_host = ova_data["vc_host"] 
        self.vc_user = ova_data["vc_user"]  
        self.vc_password = ova_data["vc_password"]  
        self.ds = ova_data["datastore"]  
        self.cluster = ova_data["cluster"]   
        self.ova_name = ova_data["ova_name"] 
        self.ova_root_password = ova_data["ova_password"]
                   
        self.count = count
        self.auth_mode=auth_mode

        if auth_mode == 'ldap_auth':
            with open(os.getcwd() + '/tests/nightly-test/configuration/ldap_auth.json') as ldap_config:
                ldap_data = json.load(ldap_config)

            self.ldap_url = ldap_data["ldap_url"] 
            self.ldap_searchdn = ldap_data["ldap_searchdn"] 
            self.ldap_search_pwd = ldap_data["ldap_search_pwd"] 
            self.ldap_filter = ldap_data["ldap_filter"] 
            self.ldap_basedn = ldap_data["ldap_basedn"] 
            self.ldap_uid = ldap_data["ldap_uid"] 
            self.ldap_scope = ldap_data["ldap_scope"] 
            self.ldap_timeout = ldap_data["ldap_timeout"] 
                
        self.harbor_password='Harbor12345'
        self.log_path=None 
        self.ip=None
        self.netmask=None
        self.gateway=None
        self.dns=None
        self.ovf_tool_path=None
        self.DEFAULT_LOCAL_OVF_TOOL_PATH = '/home/harbor-ci/ovftool/ovftool'
        self.ova_endpoints = []
        self.ova_names = []

    def __generate_ova_names(self):
        for i in range(0, self.count):
            ova_name_temp = ''
            ova_name_temp = self.ova_name +"-"+ datetime.now().isoformat().replace(":", "-").replace(".", "-")
            time.sleep(1)
            self.ova_names.append(ova_name_temp)
    
    def __set_ovf_tool(self):
        if not self.ova_endpoints:
            self.ovf_tool_path = self.DEFAULT_LOCAL_OVF_TOOL_PATH
        if not os.path.isfile(self.ovf_tool_path):
            logger.error("ovftool not found.")
        return

    def __compose_cmd(self, ova_name):
        cmd = ''

        if self.auth_mode == "db_auth":
            cmd = (
                '"%s" --X:"logFile"="./deploy_oms.log" --overwrite --powerOn --datastore=\'%s\' --noSSLVerify --acceptAllEulas --name=%s \
                --X:injectOvfEnv --X:enableHiddenProperties  --prop:root_pwd=\'%s\' --prop:permit_root_login=true --prop:auth_mode=\'%s\' \
                --prop:harbor_admin_password=\'%s\' --prop:max_job_workers=5   %s  \
                vi://%s:\'%s\'@%s/Datacenter/host/%s'
                % (self.ovf_tool_path, self.ds, ova_name,
                    self.ova_root_password, self.auth_mode,
                    self.harbor_password, self.ova_path,
                    self.vc_user, self.vc_password, self.vc_host, self.cluster
                )
            )

        if self.auth_mode == "ldap_auth":
            cmd = (
                '"%s" --X:"logFile"="./deploy_oms.log" --overwrite --powerOn --datastore=\'%s\' --noSSLVerify --acceptAllEulas --name=%s \
                --X:injectOvfEnv --X:enableHiddenProperties  --prop:root_pwd=\'%s\' --prop:permit_root_login=true --prop:auth_mode=\'%s\' \
                --prop:harbor_admin_password=\'%s\' --prop:max_job_workers=5 \
                --prop:ldap_url=\'%s\' --prop:ldap_searchdn=\'%s\' --prop:ldap_search_pwd=\'%s\' \
                --prop:ldap_filter=\'%s\' \
                --prop:ldap_basedn=\'%s\' \
                --prop:ldap_uid=\'%s\' --prop:ldap_scope=\'%s\'  --prop:ldap_timeout=\'%s\'   %s  \
                vi://%s:\'%s\'@%s/Datacenter/host/%s'
                % (self.ovf_tool_path, self.ds, ova_name,
                    self.ova_root_password, self.auth_mode,
                    self.harbor_password, 
                    self.ldap_url, self.ldap_searchdn,
                    self.ldap_search_pwd, self.ldap_filter,
                    self.ldap_basedn, self.ldap_uid,
                    self.ldap_scope, self.ldap_timeout,
                    self.ova_path,
                    self.vc_user, self.vc_password, self.vc_host, self.cluster
                )
            )
        return cmd     
   
    def deploy(self):
        self.__generate_ova_names()
        self.__set_ovf_tool()

        for i in range(0, self.count):
            cmd = self.__compose_cmd(self.ova_names[i])
            logger.info(cmd)
            try:
                subprocess.check_output(cmd, shell=True)
            except Exception, e:
                logger.info(e)
                time.sleep(5)
                # try onre more time if any failure.
                subprocess.check_output(cmd, shell=True) 
            logger.info("Successfully deployed harbor OVA.")

            ova_endpoint = ''
            ova_endpoint = govc_utils.getvmip(self.vc_host, self.vc_user, self.vc_password, self.ova_names[i])
            if ova_endpoint is not '':
                self.ova_endpoints.append(ova_endpoint)

        if self.validate(): 
            return self.ova_endpoints, self.ova_names

        return None

    def destory(self):
        for item in self.ova_names:
            govc_utils.destoryvm(self.vc_host, self.vc_user, self.vc_password, item)

    def validate(self):
        for item in self.ova_endpoints:
            is_harbor_ready = harbor_util.wait_for_harbor_ready("https://"+item)
            if not is_harbor_ready:
                logger.info("Harbor is not ready after 10 minutes.")
                return False
            logger.info("%s is ready for test now..." % item)
        return True