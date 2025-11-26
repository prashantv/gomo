#!/bin/bash
set -euo pipefail
set -x

# regenerate _site with docs
rm -rf _site
doc2go -home gomo.prashantv.com $(mise run list-module-packages)

# add meta tag for go import
find _site -iname *.html -exec go run scripts/appendmeta/appendmeta.go {} \;

