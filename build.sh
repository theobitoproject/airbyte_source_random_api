#! /bin/bash

tag=$(date +%s%N)
echo "building source random api source connector with tag: $tag"

docker build . -t bingo/source-random-api-v2:${tag}

echo "pushed ${tag}"
