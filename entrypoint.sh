#!/bin/sh

exec localgate start --port "${PORT:-9000}" --hostname "${HOSTNAME}"
