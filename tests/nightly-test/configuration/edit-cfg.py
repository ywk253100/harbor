#!/usr/bin/env python3
import argparse
import json
import os
import sys
import re
from io import StringIO

def main():
    parser = argparse.ArgumentParser(description='The script modifes harbor.cfg based on input')
    parser.add_argument('--config', '-c', dest='cfg_path', required=True, help='The path to harbor.cfg')
    parser.add_argument('--in-json', '-i', dest='json_input', required=True, help='The path to the json file input file in form of {"k":"v"} to write v to key k in harbor.cfg, if k does not appear in the cfg file as a key, it will be ignored silently.')
    args = parser.parse_args() 
    if not os.path.isfile(args.cfg_path):
        print ("The config file of Harbor does not exist, path: %s" % args.cfg_path)
        sys.exit(1)
    if not os.path.isfile(args.json_input):
        print ("The json parm file does not exist, path: %s" % args.json_input)
        sys.exit(1)
    pstr = r'\s*(\w+)\s*=.*'
    p = re.compile(pstr)
    with open(args.json_input) as f:
        d = json.load(f)
    with open(args.cfg_path, "r") as src:
        lines = src.readlines()
    with open(args.cfg_path, "w") as dst:
        for l in lines:
            m = p.match(l)
            if m is not None and m.groups()[0] in d.keys():
                k = m.groups()[0]
                print("Set attribute %s to %s"%(k, d[k]))
                l = "%s = %s\n" % (k, d[k])
            dst.write(l)

if __name__ == "__main__":
    main()
