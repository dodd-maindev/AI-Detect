#!/bin/bash
# Compile XDP firewall program to BPF bytecode
# Prerequisites: sudo apt-get install -y clang llvm libbpf-dev linux-headers-$(uname -r)

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SRC="$SCRIPT_DIR/xdp_firewall.c"
OUT="$SCRIPT_DIR/xdp_firewall.o"

echo "[*] Compiling XDP firewall: $SRC -> $OUT"
clang -O2 -g -target bpf \
    -D__TARGET_ARCH_x86 \
    -I/usr/include/x86_64-linux-gnu \
    -c "$SRC" -o "$OUT"

echo "[✓] XDP firewall compiled successfully: $OUT"
ls -la "$OUT"
