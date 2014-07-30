#!/usr/bin/env python2

"""Script to convert from YAML catalog format to JSON-based registry format."""

# Written in around 15 minutes, code sucks. Big time.

from __future__ import print_function

import collections
import json
import re
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

    registry = collections.OrderedDict([
        ('version', catalog['catalog_version']),
        ('packages', collections.OrderedDict()),
    ])

    for k in sorted(catalog.keys()):
        if k == 'catalog_version':
            continue

        pkg = catalog[k]

        name = k.replace('python2.7', 'python27')
        name = name.replace('python2.6', 'python26')

        r_pkg = registry['packages'][name] = collections.OrderedDict([
            ('version', pkg['version']),
        ])

        if type(pkg['installer']) is str:
            r_pkg['installer'] = collections.OrderedDict([
                ('kind', pkg['type']),
                ('x86', pkg['installer']),
            ])
        else:
            r_pkg['installer'] = collections.OrderedDict([
                ('kind', pkg['type']),
                ('x86', pkg['installer']['x86']),
                ('x86_64', pkg['installer']['x86_64']),
            ])

        if 'custom_arguments' in pkg:
            r_pkg['installer']['kind'] = 'as-is'
            r_pkg['installer']['options'] = {'arguments': pkg['custom_arguments'].split(' ')}

        if name == 'conemu':
            r_pkg['installer']['kind'] = 'conemu'
            del r_pkg['installer']['options']

    with open(out_file, 'w') as json_file:
        json.dump(registry, json_file, indent=4)


if __name__ == '__main__':
    main()
