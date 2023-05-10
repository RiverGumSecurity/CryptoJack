#!/usr/bin/env python3

import yaml
import collections
import argparse
import pathlib
import re
import sys


class IndentDumper(yaml.Dumper):
    def increase_indent(self, flow=False, indentless=False):
        return super(IndentDumper, self).increase_indent(flow, False)


class YamlConvert():

    yamlkeys = {}

    def __init__(self, indir, outdir):
        self.indir = pathlib.Path(indir)
        self.outdir = pathlib.Path(outdir)
        if self.indir == self.outdir:
            print('[-] Error: input dir cannot be the same as output dir')
            sys.exit(1)
        yaml.add_representer(
            collections.defaultdict,
            yaml.representer.Representer.represent_dict)
        self.run()

    def run(self):
        if not self.outdir.exists():
            self.outdir.mkdir()
        for f in self.indir.iterdir():
            if f.suffix != '.yml':
                continue
            self.process_yaml(f)
        print('UNIQUE YAML KEYS')
        print('================')
        for k in sorted(self.yamlkeys):
            print(f'-> {k}')

    def process_yaml(self, filename):
        fh = open(filename, 'rt')
        y = yaml.safe_load(fh)
        fh.close()

        newyaml = collections.defaultdict(lambda: [])
        for i in y:
            if not re.match(r'^[\w ]+$', i['ioc_type']) or \
                    i['ioc_type'] == 'md5' or \
                    i['ioc_type'].startswith('sha') or \
                    i['ioc_type'] == 'description':
                continue
            elif i['ioc_type'] == 'service' or \
                    i['ioc_type'].startswith('command_line') or \
                    i['ioc_type'] == 'scheduled task':
                newyaml['command'].append(i['data'])
                continue
            elif i['ioc_type'].startswith('ip'):
                newyaml['ip'].append(i['data'])
                continue
            elif i['ioc_type'] == 'url' or re.match(r'^https?://', i['data']):
                newyaml['web_request'].append(i['data'])
                continue
            elif i['ioc_type'].startswith('file_path') or \
                     i['ioc_type'] == 'filepath' or \
                     i['ioc_type'] == 'file_name':
                newyaml['filename'].append(i['data'])
                continue
            elif 'registry_path_key' in i['ioc_type']:
                newyaml['registry_key'].append(i['data'])
                continue
            newyaml[i['ioc_type']].append(i['data'])

        # create new file
        newfile = self.outdir / filename.name
        print(f'[*] Creating new YAML file: {newfile}')
        fh = open(newfile, 'wt')
        fh.write(yaml.dump(
            newyaml, allow_unicode=True,
            Dumper=IndentDumper))
        fh.close()

        # add to unique keys
        for k in newyaml:
            self.yamlkeys[k] = ''


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('indir', help='location of source yaml files')
    parser.add_argument('outdir', help='destination to write new files')
    args = parser.parse_args()
    YamlConvert(args.indir, args.outdir)
