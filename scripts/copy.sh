#!/bin/sh

src=$1
dest=$2
patch=13

rm -rf $dest
cp -r $src $dest
rm $dest/go.mod $dest/go.sum
find $dest -type f | xargs sed -i -e "s#k1LoW/go-github-client/$src#k1LoW/go-github-client/$dest#g"
find $dest -type f | xargs sed -i -e "s#google/go-github/$src#google/go-github/$dest#g"
find $dest -type f | grep -e '-e' | xargs rm
cd $dest
echo "module \"$(pwd | sed -e 's/.*\/src\///')\"" > go.mod
go get github.com/migueleliasweb/go-github-mock@v1.0.1
go mod tidy
go get -u golang.org/x/crypto
go get -u github.com/cloudflare/circl
go mod tidy
go test ./...
git add .
git commit -m "Update $dest"
git tag $(cat go.mod | grep "google/go-github/$dest" | cut -f 2 -d ' ' | awk -F. -v patch=$patch '{print $1 "." $2 "." $3+patch}') || true
