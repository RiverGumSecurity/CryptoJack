#!/usr/bin/env python3

import yaml
import csv
import argparse


class IOC2Yaml():

    def __init__(self, iocfile):
        self.iocfile = iocfile
        self.parse_csv()

    def parse_csv(self):
        content = []
        with open(self.iocfile, newline='', mode='rt', encoding='utf-8-sig') as f:
            iocs = csv.reader(f, delimiter=',')
            for i, r in enumerate(iocs):
                try:
                    if r[0] == '' \
                            or r[1].lower().strip().startswith('data') \
                            and r[2].lower().strip().startswith('note'):
                        continue
                    key = 'ioc{:03d}'.format(i)
                    ioc = {
                        'ioc_type': r[0].strip().lower(),
                        'data': r[1].strip(),
                        'note': r[2].strip(),
                    }
                except:
                    continue

                try:
                    extra = []
                    for i in range(3, 10):
                        if len(r[i]) > 0:
                            extra.append(r[i].strip())
                    if extra:
                        ioc['extra'] = extra
                except:
                    pass
                content.append(ioc)
        if content:
            print(yaml.dump(content))

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('iocfile', help='CSV file with IOC data')
    args = parser.parse_args()
    IOC2Yaml(args.iocfile)
