#!/bin/sh
set -eu
cd "$(dirname "$0")/../.."
test_dir=$(mktemp -d)
trap 'rm -rf "$test_dir"' EXIT
cp go.mod "$test_dir/go.mod"
cp go.sum "$test_dir/go.sum"
go mod edit -modfile="$test_dir/go.mod" -require=machine@v0.0.0 -replace="machine=$(pwd)/rtl8720dn/testdata/machine"
go test -modfile="$test_dir/go.mod" -tags=rtl8720dn_hosttest ./rtl8720dn "$@"
