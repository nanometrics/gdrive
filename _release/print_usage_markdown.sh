#!/bin/bash

echo '## Usage'
echo '```'
godrive help | tail -n+3
echo '```'

IFS=$'\n'

help=$(godrive help | grep global | sed -E 's/ \[[^]]+\]//g' | sed -E 's/ <[^>]+>//g' | sed -E 's/ {2,}.+//' | sed -E 's/^godrive //')

for args in $help; do
    cmd="godrive help $args"
    echo
    eval $cmd | sed -e '1s/^/#### /' | sed -e $'1s/$/\\\n```/' | sed -e 's/pii/<user>/'
    echo '```'
done
