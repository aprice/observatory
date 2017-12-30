#!/bin/bash
WD=$(pwd)
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
FLAGS="-cpu=1,8 -benchtime=3s -benchmem"

mkdir $DIR/reports
cd $DIR/server/
go test -run=^$ -bench=. -tags="mongo debug" $FLAGS | tee $DIR/reports/new.out
git stash save || exit 1
go test -run=^$ -bench=. -tags="mongo debug" $FLAGS | tee $DIR/reports/old.out
cd $WD
git stash pop || exit 1
benchcmp $DIR/reports/old.out $DIR/reports/new.out | tee $DIR/reports/comparison.out
