#!/bin/bash

SCRIPT_ROOT="$(cd "${0%/*}" && pwd)/"

REPO_ROOT="${SCRIPT_ROOT}../"

AGENT_DIR="${REPO_ROOT}/cmd/agent"
AGENT_BINARY_NAME="agent"
AGENT_PATH="${AGENT_DIR}/${AGENT_BINARY_NAME}"

SERVER_DIR="${REPO_ROOT}/cmd/server"
SERVER_BINARY_NAME="server"
SERVER_PATH="${SERVER_DIR}/${SERVER_BINARY_NAME}"

RUN_FUNCTIONS="${SCRIPT_ROOT}functions.bash"
source "${RUN_FUNCTIONS}"

echo "SCRIPT_ROOT=${SCRIPT_ROOT}"
echo "REPO_ROOT=${REPO_ROOT}"

eval "$(cd "$SERVER_DIR" && go build -buildvcs=false -o "$SERVER_BINARY_NAME")"
eval "$(cd "$AGENT_DIR" && go build -buildvcs=false -o "$AGENT_BINARY_NAME")"


echo "Running iter4"
iter4

echo "Running iter5"
iter5

echo "Running iter6"
iter6

echo "Running iter7"
iter7

echo "Running iter8"
iter8

echo "Running iter9"
iter9

# Do not forget to start pg database on expected port or the later tests will fail
echo "Running iter10"
iter10

echo "Running iter11"
iter11

echo "Running iter12"
iter12

echo "Running iter13"
iter13

echo "Running iter14"
iter14
