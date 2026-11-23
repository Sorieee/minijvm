#!/usr/bin/env bash
# 编译 M03 对拍用的固定测试类到 internal/classfile/testdata/。
#
# classfile 包的测试用 go:embed 内嵌这些字节，因此必须先编译再跑测试：
#   ./scripts/build-testdata.sh && go test ./internal/classfile/
#
# 要求：JDK 9+ 的 javac（用 --release 8 产出 v52 字节码）。
set -euo pipefail

cd "$(dirname "$0")/.."

if ! command -v javac >/dev/null 2>&1; then
    echo "error: javac not found in PATH" >&2
    exit 1
fi

out=internal/classfile/testdata
mkdir -p "$out"

# shellcheck disable=SC2046
javac --release 8 -encoding UTF-8 -d "$out" $(find tests/java/meta -name '*.java')

echo "OK: test data compiled to $out/"
