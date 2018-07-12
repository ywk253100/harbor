import os, subprocess
import time
import sys

from subprocess import call
import json

dir_path = os.path.dirname(os.path.realpath(__file__))
sys.path.append(dir_path + '../utils')
import harbor_util
import nlogging
logger = nlogging.create_logger(__name__)

class Executor(): 

    def __init__(self, harbor_endpoint, notary_server_endpoint, harbor_endpoint1='', harbor_pwd='Harbor12345'):
        self.harbor_endpoint = harbor_endpoint
        self.notary_server_endpoint = notary_server_endpoint
        self.harbor_endpoint1 = harbor_endpoint1
        self.harbor_user = "admin"
        self.harbor_pwd = harbor_pwd
        self.auth_mode = harbor_util.get_auth_mode(self.harbor_endpoint, self.harbor_user, self.harbor_pwd)
        self.e2e_engine = "vmware/harbor-e2e-engine:1.38"        

    def get_ts(self, auth_mode):
        with open(os.getcwd() + '/tests/nightly-test/execution/tc.json') as ts_config:
            ts_data = json.load(ts_config)
        return ts_data[auth_mode]

    def get_ca(self):
        harbor_util.get_ca(self.harbor_endpoint, self.harbor_user, self.harbor_pwd)
        if self.harbor_endpoint1 != '':
             harbor_util.get_ca(self.harbor_endpoint1, self.harbor_user, self.harbor_pwd, '/harbor/ca/ca1.crt')
    
    def __execute_test(self, cmd):
        logger.info(cmd)
        exe_result = -1
        p = subprocess.Popen(cmd, shell=True, stderr=subprocess.PIPE)
        while True:
            out = p.stderr.read(1)
            if out == '' and p.poll() != None:
                break
            if out != '':
                sys.stdout.write(out)
                sys.stdout.flush()
        exe_result = p.returncode
        return exe_result

    def __prepare(self):
        if self.auth_mode == 'db_auth':
            os.system(os.getcwd() + '/tests/nightly-test/shellscript/prepare.sh %s' % self.harbor_endpoint)
        self.get_ca()

    def execute(self):
        cmd = ''
        
        cmd_base = "docker run -i --privileged -v %s:/drone -v /harbor/ca:/ca -w /drone %s " % (os.getcwd(), self.e2e_engine)
        cmd_pybot = "pybot -v ip:%s -v notaryServerEndpoint:%s -v ip1:%s -v HARBOR_PASSWORD:%s " % (self.harbor_endpoint, self.notary_server_endpoint, self.harbor_endpoint1, self.harbor_pwd)
        cmd = cmd_base + cmd_pybot
        
        # any test execution will be setup + common + auth_mode specific + teardown.
        cmd = cmd + self.get_ts("setup") + " "
        cmd = cmd + self.get_ts("common") + " "
        if self.harbor_endpoint1 != '':    
            cmd = cmd + self.get_ts('replication') + " "
        else:
            cmd = cmd + self.get_ts(self.auth_mode) + " "
        cmd = cmd + self.get_ts("teardown") + " "

        self.__prepare()
        return self.__execute_test(cmd)