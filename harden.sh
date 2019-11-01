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

mkdir -p "$1"
cp -f {chunkPool.go,deque.go,benchmark_test.go} "$1"
[[ $? -ne 0 ]] && exit 1

perl -pi -e "s/^package deque$/package $2/g" "$1"/*
[[ $? -ne 0 ]] && exit 1
perl -pi -e "s/type deque struct/type Deque struct/g" "$1"/*
[[ $? -ne 0 ]] && exit 1
perl -pi -e "s/func \(dq \*deque\) /func (dq *Deque) /g" "$1"/*
[[ $? -ne 0 ]] && exit 1
perl -pi -e "s/func NewDeque\(\) Deque {/func NewDeque() *Deque {/g" "$1"/*
[[ $? -ne 0 ]] && exit 1
perl -pi -e "s/dq := &deque{/dq := &Deque{/g" "$1"/*
[[ $? -ne 0 ]] && exit 1

elemType="$3"
if [[ "$elemType" == "" ]]; then
    elemType='interface{}'
fi
if ! [[ -f "$1"/elem.go ]]; then
    cat <<EOF> "$1"/elem.go
package $2

type Elem = $elemType
EOF
fi
