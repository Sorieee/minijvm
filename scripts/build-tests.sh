#!/usr/bin/env bash
# 编译 tests/java 下所有 Java 测试程序到 tests/classes/。
#
# 要求：JDK 21+ 的 javac（用 --release 8 产出 v52 字节码，
# 且字符串拼接走 StringBuilder 而非 invokedynamic）。
set -euo pipefail

cd "$(dirname "$0")/.."

if ! command -v javac >/dev/null 2>&1; then
    echo "error: javac not found in PATH" >&2
    exit 1
fi

mkdir -p tests/classes

# shellcheck disable=SC2046
javac --release 8 -encoding UTF-8 -d tests/classes \
    $(find tests/java -name '*.java' -not -path '*/classes/*')

echo "OK: test classes compiled to tests/classes/"
