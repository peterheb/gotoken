import urllib.request
import gzip

# This script downloads the pae-enwiki-2023-04-1gb.txt file from cloud storage.
# It's pretty large (1GB), so is not included in the git repo. This file is not
# required by the regular unit tests, but can be used for benchmarking or
# testing via manual code changes. See the commented out lines in the various
# tokenizer_test.go files and gen_ground_truth.py to generate test files.
#
# The script that generated the contents of this file can be found here:
#
#    https://gist.github.com/peterheb/f672cb7c754fa16f8d0a0155d2dc6db2
#
# (the header lines were created by hand)
#
# For the curious, a larger 7GiB extract can be retrieved from:
#
# https://gotoken.phebert.dev/pae-enwiki-2023-04.txt.gz
#
# Note that one of the gotoken tests does fail in the larger dataset
# with cl100k_base, due to outdated Unicode 13.0 data in Go.

url = 'https://gotoken.phebert.dev/pae-enwiki-2023-04-1gb.txt.gz'
filename = 'pae-enwiki-2023-04-1gb.txt'

# Cloudflare (which hosts gotoken.phebert.dev) rejects the download as coming
# from an evil bot without some headers in the request. Some debugging output is
# enabled in case, at some future date, these headers are not "enough".
headers = {'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8',
           'User-Agent': 'get_wiki_1gb/1.0 (https://github.com/peterheb/gotoken)'}
req = urllib.request.Request(url, headers=headers)
http_handler = urllib.request.HTTPSHandler(debuglevel=1)
opener = urllib.request.build_opener(http_handler)
urllib.request.install_opener(opener)

try:
    with urllib.request.urlopen(req) as response, gzip.GzipFile(fileobj=response) as uncompressed, open(filename, 'wb') as out_file:
        while True:
            data = uncompressed.read(1024 * 1024)
            if not data:
                break
            out_file.write(data)
except urllib.error.HTTPError as e:
    print(f"HTTP Error {e.code}: {e.reason}")
    print(e.headers)
    print(e.read())
