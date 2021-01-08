
PROJECTNAME=$(shell basename "$(PWD)")

GOBIN = ./build/bin

exec:
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go build -o ./build/bin/tpsmonitor main.go
	@echo "Done building $(PROJECTNAME)."
	@echo "Run \"$(GOBIN)/tpsmonitor\" to launch tpsmonitor."



