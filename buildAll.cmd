go build -o zvart.exe cmd/zvartconsole/main.go 
go build -gcflags=all="-N -l" -o zvartDbg.exe cmd/zvartconsole/main.go
@pause
