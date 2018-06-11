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
        self.init_from_input()

    @staticmethod
    def parse_input():
        parser = argparse.ArgumentParser(description='run testcase') 
        parser.add_argument('--endpoint', '-e', dest='endpoint', required=True, help='The endpoint of harbor.')

        args = parser.parse_args()
        return (args.endpoint)

    def init_from_input(self):
        (self.endpoint) = Parameters.parse_input()


def main():
    commandline_input = Parameters()
    
    try:
        tc_executor = Executor(commandline_input.endpoint)
        tc_executor.execute()

    except Exception, e:
        print e
        sys.exit(1)

if __name__ == '__main__':
    main()
