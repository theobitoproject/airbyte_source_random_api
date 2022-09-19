#! /bin/bash

tag=$(date +%s%N)
echo "building random api source connector with tag: $tag"

docker build . -t bingo/random-api:${tag}

echo "pushed ${tag}"
