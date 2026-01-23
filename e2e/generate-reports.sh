#!/usr/bin/env bash
# This script is a convenience for generating the migration report for every test case.
set -euo pipefail

BYO_AGC="/subscriptions/11111111-1111-1111-bbbb-bbbbbbbbbbbb/resourceGroups/onebox-rg/providers/Microsoft.ServiceNetworking/trafficControllers/test-tc"

pushd "$(dirname "$(go env GOMOD)")"

gen_report() {
    dir="$1"
    inputs=("$dir"/input/*.yaml)
    report_file="$dir/report.yaml"
    go run ./cmd files "${inputs[@]}" --dry-run  --byo-resource-id "$BYO_AGC" | awk '/# Migration Report/{flag=1; next} flag' > "$report_file"
}

test_dirs=$(find e2e/ -type d -name input -exec dirname {} \;)

for dir in $test_dirs; do
    gen_report "$dir"
done