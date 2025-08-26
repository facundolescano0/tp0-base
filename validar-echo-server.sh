#!/bin/bash
set -e

MENSAJE="hola-mundo"

RESPUESTA=$(docker run --rm --network tp0_testing_net busybox sh -c "echo $MENSAJE | nc server 12345")

if [ "$RESPUESTA" = "$MENSAJE" ]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
    exit 1
fi