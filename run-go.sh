#!/bin/bash

# Clean and compile Go code before running it here to make sure
# we don't run OSX binaries in the container
make goclean
make go

exec $@
