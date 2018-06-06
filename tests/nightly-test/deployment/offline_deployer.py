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
logger = nlogging.create_logger(__name__)

class OfflineDeployer():

    def __init__(self, auth_mode, url):
        self.auth_mode = auth_mode
        self.ip = self.__get_ip()
        self.url = url        

    def deploy(self):
        try:
            SHELL_SCRIPT_DIR = os.getcwd() + '/tests/nightly-test/shellscript/'
            cmd = SHELL_SCRIPT_DIR + "/offline_installer.sh %s %s %s" % (self.auth_mode, self.ip, self.url)
            os.system(cmd)
            if self.validate():
                return self.ip
            return None
        except Exception as e:
            logger.info("Caught Exception When To Deploy offline installer : " + str(e))

    def destory(self):
        ## clean env
        pass
    
    def validate(self):
        is_harbor_ready = harbor_util.wait_for_harbor_ready("https://"+self.ip)
        if not is_harbor_ready:
            logger.info("Harbor is not ready after 10 minutes.")
            return False
        logger.info("%s is ready for test now..." % self.ip)
        return True
    
    def __get_ip(self):
        try:
            arg='ip route list'    
            p=subprocess.Popen(arg,shell=True,stdout=subprocess.PIPE)
            data = p.communicate()
            sdata = data[0].split()
            ip = sdata[ sdata.index('src')+1 ]
            return ip 
        except Exception as e:
            logger.info("Caught Exception on getting ip : " + str(e))