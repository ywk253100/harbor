#!/usr/bin/python

import os
import sys
from optparse import OptionParser
dir_path = os.path.dirname(os.path.realpath(__file__))
sys.path.append(dir_path + '/utils')
sys.path.append(dir_path + '/deployment')
import nlogging
from offline_deployer import *
from ova_deployer import *
logger = nlogging.create_logger(__name__)
import argparse

class Parameters(object):
    def __init__(self):
        self.installer_type = ''
        self.url = ''
        self.auth_mode = ''
        self.init_from_input()

    @staticmethod
    def parse_input():
        parser = argparse.ArgumentParser(description='installer of harbor') 
        parser.add_argument('--installer-type', '-i', dest='installer_type', required=True, help='The installer type: offline or ova.')
        parser.add_argument('--url', '-u', dest='url', required=True, help='The url to dowoload the installer.')
        parser.add_argument('--auth-mode', '-a', dest='auth_mode', required=True, help='')

        args = parser.parse_args()
        return (args.installer_type, args.url, args.auth_mode)

    def init_from_input(self):
        (self.installer_type, self.url, self.auth_mode) = Parameters.parse_input()


def main():
    commandline_input = Parameters()
    
    try:
        if commandline_input.installer_type == 'ova':
            ova_deployer = OVADeployer(commandline_input.auth_mode, commandline_input.url)
            logger.info("Going to deploy harbor ova..")
            harbor_endpoint = ova_deployer.deploy()
            print harbor_endpoint 
        elif commandline_input.installer_type == 'offline':
            offline_deployer = OfflineDeployer(commandline_input.auth_mode, commandline_input.url)
            logger.info("Going to deploy harbor offline..")
            harbor_endpoint = offline_deployer.deploy()
            print harbor_endpoint          

    except Exception, e:
        print e
        sys.exit(1)

if __name__ == '__main__':
    main()
