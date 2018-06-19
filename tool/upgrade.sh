#!/bin/bash

SRC_REPO="git@bitbucket.org:cloudwallet/litereplica.git"
SRC_DIR=litereplica

function update {
    diff -q $1 $2 && echo "no update in $1" || cp $1 $2
}

function clone {
    git clone $SRC_REPO $SRC_DIR
    cd $SRC_DIR
    git checkout go-sqlite3
}

if [ -d "$SRC_DIR" ]; then
    cd $SRC_DIR
    REPO=`git config remote.origin.url`
    if [ "$REPO" == "$SRC_REPO" ]; then
        git checkout go-sqlite3
        git pull origin go-sqlite3
    else
        cd ..
        rm -rf $SRC_DIR
        clone
    fi
else
    clone
fi

update sqlite3.c ../../sqlite3-binding.c
update sqlite3.h ../../sqlite3-binding.h
