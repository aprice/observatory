#!/bin/bash
WD=$(pwd)
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
mkdir $DIR/reports
git checkout HEAD^
cd $DIR/server/ && go test -run=^$ -bench=. -cpu=1,8 -benchtime=2s -benchmem -tags="mongo debug" | tee old.out
git checkout HEAD
cd $DIR/server/ && go test -run=^$ -bench=. -cpu=1,8 -benchtime=2s -benchmem -tags="mongo debug" | tee new.out
benchcmp $DIR/server/old.out $DIR/server/new.out | tee $DIR/reports/comparison.out
cd $WD
