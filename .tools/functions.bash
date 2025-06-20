#!/bin/bash

declare SERVER_PORT
declare ADDRESS
declare TEMP_FILE
    
SERVER_PORT=$(./random unused-port)
ADDRESS="localhost:${SERVER_PORT}"
TEMP_FILE=$(./random tempfile)

iter4() {
    ./metricstest -test.v -test.run=^TestIteration4$ \
        -agent-binary-path="${AGENT_PATH}" \
        -binary-path="${SERVER_PATH}" \
        -server-port="${SERVER_PORT}" \
        -source-path="${REPO_ROOT}"
}

iter5() {
    ./metricstest -test.v -test.run=^TestIteration5$ \
        -agent-binary-path="${AGENT_PATH}" \
        -binary-path="${SERVER_PATH}" \
        -server-port="${SERVER_PORT}" \
        -source-path="${REPO_ROOT}"
}


iter6() {
    ./metricstest -test.v -test.run=^TestIteration6$ \
        -agent-binary-path="${AGENT_PATH}" \
        -binary-path="${SERVER_PATH}" \
        -server-port="${SERVER_PORT}" \
        -source-path="${REPO_ROOT}"
}


iter7() {
    ./metricstest -test.v -test.run=^TestIteration7$ \
        -agent-binary-path="${AGENT_PATH}" \
        -binary-path="${SERVER_PATH}" \
        -server-port="${SERVER_PORT}" \
        -source-path="${REPO_ROOT}"
}

iter8() {
    ./metricstest -test.v -test.run=^TestIteration8$ \
        -agent-binary-path="${AGENT_PATH}" \
        -binary-path="${SERVER_PATH}" \
        -server-port="${SERVER_PORT}" \
        -source-path="${REPO_ROOT}"
}

iter9() {
    ./metricstest -test.v -test.run=^TestIteration9$ \
        -agent-binary-path="${AGENT_PATH}" \
        -binary-path="${SERVER_PATH}" \
        -server-port="${SERVER_PORT}" \
        -file-storage-path="${TEMP_FILE}" \
        -server-port="${SERVER_PORT}" \
        -source-path="${REPO_ROOT}"
}

iter10() {
    ./metricstest -test.v -test.run=^TestIteration10[AB]$ \
        -agent-binary-path="${AGENT_PATH}" \
        -binary-path="${SERVER_PATH}" \
        -server-port="${SERVER_PORT}" \
        -file-storage-path="${TEMP_FILE}" \
        -database-dsn='postgres://go-metrics@localhost:15432/go-metrics' \
        -server-port="${SERVER_PORT}" \
        -source-path="${REPO_ROOT}"
}

iter11() {
    ./metricstest -test.v -test.run=^TestIteration11$ \
        -agent-binary-path="${AGENT_PATH}" \
        -binary-path="${SERVER_PATH}" \
        -server-port="${SERVER_PORT}" \
        -file-storage-path="${TEMP_FILE}" \
        -database-dsn='postgres://go-metrics@localhost:15432/go-metrics' \
        -server-port="${SERVER_PORT}" \
        -source-path="${REPO_ROOT}"
}

iter12() {
    ./metricstest -test.v -test.run=^TestIteration12$ \
        -agent-binary-path="${AGENT_PATH}" \
        -binary-path="${SERVER_PATH}" \
        -server-port="${SERVER_PORT}" \
        -file-storage-path="${TEMP_FILE}" \
        -database-dsn='postgres://go-metrics@localhost:15432/go-metrics' \
        -server-port="${SERVER_PORT}" \
        -source-path="${REPO_ROOT}"
}

iter12() {
    ./metricstest -test.v -test.run=^TestIteration12$ \
        -agent-binary-path="${AGENT_PATH}" \
        -binary-path="${SERVER_PATH}" \
        -server-port="${SERVER_PORT}" \
        -file-storage-path="${TEMP_FILE}" \
        -database-dsn='postgres://go-metrics@localhost:15432/go-metrics' \
        -server-port="${SERVER_PORT}" \
        -source-path="${REPO_ROOT}"
}

iter13() {
    ./metricstest -test.v -test.run=^TestIteration13$ \
        -agent-binary-path="${AGENT_PATH}" \
        -binary-path="${SERVER_PATH}" \
        -server-port="${SERVER_PORT}" \
        -file-storage-path="${TEMP_FILE}" \
        -database-dsn='postgres://go-metrics@localhost:15432/go-metrics' \
        -server-port="${SERVER_PORT}" \
        -source-path="${REPO_ROOT}"
}

iter14() {
    ./metricstest -test.v -test.run=^TestIteration14$ \
        -agent-binary-path="${AGENT_PATH}" \
        -binary-path="${SERVER_PATH}" \
        -database-dsn='postgres://go-metrics@localhost:15432/go-metrics' \
        -key="${TEMP_FILE}" \
        -server-port="${SERVER_PORT}" \
        -source-path="${REPO_ROOT}"
}
