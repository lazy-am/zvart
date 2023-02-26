#!/bin/bash

echo "Build release"
go build -o zvart cmd/zvartconsole/main.go
echo "Build debug"
go build -gcflags=all="-N -l" -o zvartDbg cmd/zvartconsole/main.go
echo "End"


