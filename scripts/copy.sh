#!/bin/sh

src=$1
dest=$2
patch=17

rm -rf $dest
cp -r $src $dest
rm $dest/go.mod $dest/go.sum
find $dest -type f | xargs sed -i -e "s#k1LoW/go-github-client/$src#k1LoW/go-github-client/$dest#g"
find $dest -type f | xargs sed -i -e "s#google/go-github/$src#google/go-github/$dest#g"
find $dest -type f | grep -e '-e' | xargs rm
cd $dest
echo "module \"$(pwd | sed -e 's/.*\/src\///')\"" > go.mod
go mod tidy
go get -u golang.org/x/crypto
go get -u golang.org/x/net
go get -u github.com/cloudflare/circl
go get -u google.golang.org/protobuf
go get -u github.com/golang-jwt/jwt/v4
go get -u
go mod tidy
go test ./...
git add .
git commit -m "Update $dest"
git tag $(cat go.mod | grep "google/go-github/$dest" | cut -f 2 -d ' ' | awk -F. -v patch=$patch '{print $1 "." $2 "." $3+patch}') || true
