#!/usr/bin/python

import os
import sys
from optparse import OptionParser
dir_path = os.path.dirname(os.path.realpath(__file__))
sys.path.append(dir_path + '/deployment')
sys.path.append(dir_path + '../utils')
import nlogging
import deployer
logger = nlogging.create_logger(__name__)

class Parameters(object):
    def __init__(self):
        self.installer_type = ''
        self.url = ''
        self.auth_mode = ''
        self.init_from_input()

    @staticmethod
    def parse_input():
        usage = "usage: %prog [options] <installer> <url> <auth_mode>"
        parser = OptionParser(usage)
        parser.add_option('--installer-type', '-i', dest='installer_type', required=True, help='The installer type: offline or ova.')
        parser.add_option('--url', '-u', dest='url', required=True, help='The url to dowoload the installer.')
        parser.add_option('--auth_mode', '-u', dest='auth_mode', required=True, help='')

        (options, args) = parser.parse_args()
        return (options.installer_type, options.url, options.auth_mode)

    def init_from_input(self):
        (self.installer_type, self.url, self.auth_mode) = Parameters.parse_input()


def main():
    commandline_input = Parameters()
    
    try:
        if commandline_input.installer_type == 'ova':
            pass
        elif commandline_input.installer_type == 'offline':
            offline_deployer = OfflineDeployer(commandline_input.auth_mode, commandline_input.url)
            logger.info("Going to deploy harbor offline..")
            harbor_endpoints = offline_deployer.deploy()          

    except Exception, e:
        print e
        sys.exit(1)

if __name__ == '__main__':
    main()
