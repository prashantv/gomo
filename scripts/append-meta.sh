#!/bin/bash
set -euo pipefail
set -x

perl -i -pe 's{<head(.*?)>}{<head$1>\n<meta name="go-import" content="gomo.prashantv.com/ git https://github.com/prashantv/go">
}i' $1

