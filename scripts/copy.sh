#!/bin/sh

src=$1
dest=$2

rm -rf $dest
cp -r $src $dest
rm $dest/go.mod $dest/go.sum
find $dest -type f | xargs sed -i -e "s#k1LoW/go-github-client/$src#k1LoW/go-github-client/$dest#g"
find $dest -type f | xargs sed -i -e "s#google/go-github/$src#google/go-github/$dest#g"
cd $dest
echo "module \"$(pwd | sed -e 's/.*\/src\///')\"" > go.mod
go mod tidy
go test ./...
git tag $(cat go.mod | grep "google/go-github/$dest" | cut -f 2 -d ' ') -f
