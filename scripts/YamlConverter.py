#!/usr/bin/env python3

import yaml
import collections
import argparse
import pathlib
import re


class IndentDumper(yaml.Dumper):
    def increase_indent(self, flow=False, indentless=False):
        return super(IndentDumper, self).increase_indent(flow, False)


class YamlConvert():

    newyaml = collections.defaultdict(lambda: [])
    yamlkeys = {}

    def __init__(self, yamldir):
        self.yamldir = yamldir
        yaml.add_representer(
            collections.defaultdict,
            yaml.representer.Representer.represent_dict)
        self.run()

    def run(self):
        for f in pathlib.Path(self.yamldir).iterdir():
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

        for i in y:
            if not re.match(r'^[\w ]+$', i['ioc_type']) or \
                    i['ioc_type'] == 'md5' or \
                    i['ioc_type'].startswith('sha') or \
                    i['ioc_type'] == 'description':
                continue
            elif i['ioc_type'] == 'service' or i['ioc_type'] == 'scheduled task':
                self.newyaml['command'].append(i['data'])
                continue
            elif i['ioc_type'].startswith('ip'):
                self.newyaml['ip'].append(i['data'])
                continue
            elif i['ioc_type'] == 'url':
                self.newyaml['web_request'].append(i['data'])
                continue
            elif i['ioc_type'].startswith('file_path') or \
                     i['ioc_type'] == 'filepath' or \
                     i['ioc_type'] == 'file_name':
                self.newyaml['filename'].append(i['data'])
                continue
            elif 'registry_path_key' in i['ioc_type']:
                self.newyaml['registry_key'].append(i['data'])
                continue
            self.newyaml[i['ioc_type']].append(i['data'])

        # create new file
        newfile = pathlib.Path(pathlib.Path.cwd()) / filename.name
        if newfile != filename:
            print(f'[*] Creating new YAML file: {newfile}')
            fh = open(newfile, 'wt')
            fh.write(yaml.dump(
                self.newyaml,
                allow_unicode=True,
                Dumper=IndentDumper))
            fh.close()

        # add to unique keys
        for k in self.newyaml:
            self.yamlkeys[k] = ''


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument(
        'yamldir',
        help='directory of yaml files')
    args = parser.parse_args()
    YamlConvert(args.yamldir)
