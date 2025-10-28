#!/bin/bash

GALLICORPORA_PATH="/Users/mia/dev/personal/gallicorpora"
KRAKEN_VERSION=$(pip list 2>/dev/null | grep -i '^kraken[[:space:]]' | awk '{print $2}')
python /Users/mia/dev/personal/gallicorpora/scripts/alto2tei.py --config config.yml --version 5.3.0 --header --sourcedoc --body

