#!/usr/bin/env bash

[[ "$TRACE" ]] && set -x
pushd `dirname "$0"` > /dev/null
trap __EXIT EXIT

colorful=false
tput setaf 7 > /dev/null 2>&1
if [[ $? -eq 0 ]]; then
    colorful=true
fi

function __EXIT() {
    popd > /dev/null
}

function printError() {
    $colorful && tput setaf 1
    >&2 echo "Error: $@"
    $colorful && tput setaf 7
}

function printImportantMessage() {
    $colorful && tput setaf 3
    >&2 echo "$@"
    $colorful && tput setaf 7
}

function printUsage() {
    $colorful && tput setaf 3
    >&2 echo "$@"
    $colorful && tput setaf 7
}

if [[ $# -lt 2 ]]; then
    printUsage "Usage: ./`basename $0` <outputDir> <packageName> [elemType]"
    exit 1
fi

if [ "${1:0:1}" != "/" ]; then
    printError "outputDir must be an absolute directory"
    exit 1
fi

elemType="$3"
if [[ "$3" == *'/'* ]] && [[ "$3" == *'.'* ]]; then
    elemPackage=$(echo "$3" | perl -pe 's/(.+)\.[^\/]+/$1/g')
    elemType=$(echo "$3" | perl -pe 's/.+\/([^\/]+)/$1/g')
fi
if [[ "$elemType" == "" ]]; then
    elemType='interface{}'
fi

mkdir -p "$1"
[[ $? -ne 0 ]] && exit 1

# 1
cat <<EOF> "$1"/deque.go
package $2

import (
    "fmt"
    "sync"
    "sync/atomic"

    "github.com/edwingeng/deque"
EOF
[[ $? -ne 0 ]] && exit 1

# 2
if [[ "$elemPackage" != '' ]]; then
    echo "\"$elemPackage\"" >> "$1"/deque.go
    [[ $? -ne 0 ]] && exit 1
fi

# 3
cat <<EOF>> "$1"/deque.go
)

type Elem = $elemType

var (
    _ = deque.NumChunksAllocated
)
EOF
[[ $? -ne 0 ]] && exit 1

cat chunkPool.go | perl -0pe 's/^package.+^import \(.+?\)//gms' >> "$1"/deque.go
[[ $? -ne 0 ]] && exit 1
cat deque.go | perl -0pe 's/^package.+^import \(.+?\)//gms' >> "$1"/deque.go
[[ $? -ne 0 ]] && exit 1

perl -pi -e "s/^package deque$/package $2/g" "$1"/deque.go
[[ $? -ne 0 ]] && exit 1
perl -pi -e "s/type deque struct/type Deque struct/g" "$1"/deque.go
[[ $? -ne 0 ]] && exit 1
perl -pi -e "s/func \(dq \*deque\) /func (dq *Deque) /g" "$1"/deque.go
[[ $? -ne 0 ]] && exit 1
perl -pi -e "s/func NewDeque\(\) Deque {/func NewDeque() *Deque {/g" "$1"/deque.go
[[ $? -ne 0 ]] && exit 1
perl -pi -e "s/dq := &deque{/dq := &Deque{/g" "$1"/deque.go
[[ $? -ne 0 ]] && exit 1

gofmt -w "$1"/deque.go
