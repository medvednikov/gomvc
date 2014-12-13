#!/bin/sh

cd examples/quickstart
echo "1"
killall ezweb_quickstart
echo "2"
go build -o ezweb_quickstart quickstart.go && ./ezweb_quickstart  > /tmp/out.log 2>&1 &

cd ../..
go test -v


#go test controllers/*.go

