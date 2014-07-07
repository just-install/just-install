#!/usr/bin/env python2

from __future__ import print_function

import json
import sys
import yaml


def main():
    if len(sys.argv) != 3:
        print('Usage: yml2json.py in-file out-file')

        sys.exit(1)

    in_file = sys.argv[1]
    out_file = sys.argv[2]

    with open(in_file, 'r') as yaml_file:
        catalog = yaml.load(yaml_file)

    with open(out_file, 'w') as json_file:
        json.dump(catalog, json_file, indent=4, sort_keys=True)


if __name__ == '__main__':
    main()
