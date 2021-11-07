#!/bin/sh

ver=$1

rm -f $ver/go.*
cd $ver/
echo "module \"$(pwd | sed -e 's/.*\/src\///')\"" > go.mod
go mod tidy
go test ./...
git add .
git tag $(cat go.mod | grep "google/go-github/$ver" | cut -f 2 -d ' ') -f
