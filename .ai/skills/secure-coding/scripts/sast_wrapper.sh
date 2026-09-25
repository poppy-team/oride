#!/usr/bin/env sh
set -e
TARGET="."
echo "Running Static Application Security Testing (SAST) via Prumo Tooling..."
prumo tool scan-secrets ""
prumo tool check-subprocesses ""
prumo tool check-bare-errors ""
