# this makefile used in console environment.
# copy this file to project base directory.

# SET BIN NAME BY USER
binName:=pgo-demo

goBin:=go
glideBin:=glide

######## DO NOT CHANGE THE FLOWING CONTENT ########

# absolute path of makefile
mkPath:=$(abspath $(firstword $(MAKEFILE_LIST)))

# absolute base directory of project
baseDir:=$(strip $(patsubst %/, %, $(dir $(mkPath))))

binDir:=$(baseDir)/bin
srcDir:=$(baseDir)/src

.PHONY: start stop build update pgo init

start: build
	$(binDir)/$(binName)

stop:
	-killall $(binName)

build:
	export GOPATH=$(baseDir) && $(goBin) build -o $(binDir)/$(binName) $(srcDir)/Main/main.go

update:
	export GOPATH=$(baseDir) && cd $(srcDir) && $(glideBin) update

install:
	export GOPATH=$(baseDir) && cd $(srcDir) && $(glideBin) install

pgo:
	export GOPATH=$(baseDir) && cd $(srcDir) && $(glideBin) get github.com/pinguo/pgo

init:
	[ -d $(baseDir)/conf ] || mkdir $(baseDir)/conf
	[ -d $(srcDir) ] || mkdir $(srcDir)
	[ -d $(srcDir)/Command ] || mkdir $(srcDir)/Command
	[ -d $(srcDir)/Controller ] || mkdir $(srcDir)/Controller
	[ -d $(srcDir)/Lib ] || mkdir $(srcDir)/Lib
	[ -d $(srcDir)/Main ] || mkdir $(srcDir)/Main
	[ -d $(srcDir)/Model ] || mkdir $(srcDir)/Model
	[ -d $(srcDir)/Service ] || mkdir $(srcDir)/Service
	[ -d $(srcDir)/Struct ] || mkdir $(srcDir)/Struct
	[ -d $(srcDir)/Test ] || mkdir $(srcDir)/Test
	[ -f $(srcDir)/glide.yaml ] || (cd $(srcDir) && echo Y | $(glideBin) init)

help:
	@echo "make start       build and start $(binName)"
	@echo "make stop        stop process $(binName)"
	@echo "make build       build $(binName)"
	@echo "make update      glide update packages recursively"
	@echo "make install     glide install packages in glide.lock"
	@echo "make pgo         glide get pgo"
	@echo "make init        init project"
