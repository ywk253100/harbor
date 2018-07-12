#!/usr/bin/python

import os
import sys
from optparse import OptionParser
dir_path = os.path.dirname(os.path.realpath(__file__))
sys.path.append(dir_path + '/utils')
sys.path.append(dir_path + '/execution')
from test_executor import *
logger = nlogging.create_logger(__name__)
import argparse

class Parameters(object):
    def __init__(self):
        self.endpoint = ''
        self.notaryServerEndpoint = ''
        # only for replication
        self.endpoint1 = ''
        self.init_from_input()

    @staticmethod
    def parse_input():
        parser = argparse.ArgumentParser(description='run testcase') 
        parser.add_argument('--endpoint', '-e', dest='endpoint', required=True, help='The endpoint of harbor.')
        parser.add_argument('--notary-server-endpoint', '-nse', dest='notaryServerEndpoint', required=False, help='The endpoint of Notary server.')
        parser.add_argument('--endpoint1', '-e1', dest='endpoint1', required=False, help='The endpoint of harbor.')

        args = parser.parse_args()
        return (args.endpoint, args.notaryServerEndpoint, args.endpoint1)

    def init_from_input(self):
        (self.endpoint, self.notaryServerEndpoint, self.endpoint1) = Parameters.parse_input()


def main():
    commandline_input = Parameters()
    
    try:
        tc_executor = Executor(commandline_input.endpoint, commandline_input.notaryServerEndpoint, commandline_input.endpoint1)
        tc_executor.execute()

    except Exception, e:
        print e
        sys.exit(1)

if __name__ == '__main__':
    main()
