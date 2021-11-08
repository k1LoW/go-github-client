#!/bin/sh

ver=$1
patch=2

rm -f $ver/go.*
cd $ver/
echo "module \"$(pwd | sed -e 's/.*\/src\///')\"" > go.mod
go mod tidy
go test ./...
git add .
git tag -f $(cat go.mod | grep "google/go-github/$ver" | cut -f 2 -d ' ' | awk -F. -v patch=$patch '{print $1 "." $2 "." $3+patch}')
