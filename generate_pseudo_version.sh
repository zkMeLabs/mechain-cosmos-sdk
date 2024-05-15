#!/bin/bash

commit_hash=$(git rev-parse HEAD)
short_commit_hash=${commit_hash:0:12}

timestamp=$(git show -s --format=%ct HEAD)

if [[ "$OSTYPE" == "darwin"* ]]; then

    formatted_timestamp=$(date -u -r "$timestamp" +'%Y%m%d%H%M%S')
else

    formatted_timestamp=$(date -u -d @"$timestamp" +'%Y%m%d%H%M%S')
fi

pseudo_version="v0.0.0-$formatted_timestamp-$short_commit_hash"

echo "Generated pseudo-version: $pseudo_version"
