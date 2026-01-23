#!/bin/bash
# shellcheck disable=SC2046
go test -coverprofile=coverage.out $(go list ./... | grep -v -E '/(crds|e2e|testutil)/|/cmd$') && go tool cover -func=coverage.out | grep total | awk '{print $1, $NF}'
