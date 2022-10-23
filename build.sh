#! /bin/bash

tag=$(date +%s%N)
echo "building random api source connector with tag: $tag"

docker build . -t bingo/random-api-v2:${tag}

echo "pushed ${tag}"
