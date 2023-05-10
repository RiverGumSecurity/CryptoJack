#!/usr/bin/env python3

import argparse
import glob
import pathlib
import base64


class EncryptYAML():

    def __init__(self, directory):
        self.directory = directory
        self.run()

    def run(self):
        for f in glob.glob(str(pathlib.Path(self.directory) / '*.yml')):
            newfile = f + '.enc'
            print('[*] Encrypting {} -> {}'.format(f,newfile))
            with open(f, 'rb') as fh:
                content = fh.read()
            encrypted = self.xor(content, b'\xde\xad\xbe\xef')
            with open(newfile, 'wb') as fh:
                fh.write(encrypted)

    def xor(self, buf, k):
        res = b''
        for i, ch in enumerate(buf):
            res += (ch ^ k[i % len(k)]).to_bytes(1, 'little')
        return res

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('directory', help='YAML file directory')
    args = parser.parse_args()
    EncryptYAML(args.directory)
