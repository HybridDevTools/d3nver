ifndef VERBOSE
.SILENT:
endif

PACKAGE        := denver
DATE           := $(shell date +%s)
ifeq ($(wildcard .release),)
	VERSION        := $(shell git --no-pager log --pretty=format:'%h' -n 1)
else
	VERSION        := $(shell cat .release)
endif

SHELL          := bash

BUILDER_IMG    := golang:stretch
BUILDER_STATUS := $(shell docker inspect -f '{{.State.Running}}' denver_builder 2> /dev/null)

GO             := go
BASE           := $(shell pwd)

DIST           := {linux,darwin,windows}

S3BUCKET       := s3.d3nver.io/app
S3PATH         := https://s3-eu-west-1.amazonaws.com/$(S3BUCKET)

UID            := $(shell id -u)
GID            := $(shell id -g)

V  = 0
Q  = $(if $(filter 1,$V),,@)
M  = $(shell printf "\033[34;94m▶\033[0m")
M2 = $(shell printf "  \033[34;35m▶\033[0m")

########################################################
### Stages                                           ###
########################################################

help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

test: install-tools ; $(info $(M) Testing sources...) @  ## Test
	cd ./src && go test -race ./...

lint: install-tools ; $(info $(M) Linting sources...) @  ## Code lint
		$Q cd ./src && fgt golint ./...
		$Q cd ./src && fgt go vet ./...
		$Q cd ./src && fgt go fmt ./...
		$Q cd ./src && fgt goimports -w .
		$Q cd ./src && fgt errcheck -ignore Close  ./...

dev-start: ; $(info $(M) Starting dev tools...) @  ## Starting dev tools
ifeq ($(BUILDER_STATUS), true)
		$(info $(M2) $(PACKAGE)_builder already running )
else
		$(info $(M2) $(PACKAGE)_builder not running, starting ... )
		docker run --rm -it --detach --name $(PACKAGE)_builder --volume $$PWD:/go/src/$(PACKAGE) $(BUILDER_IMG)
endif

dev-stop: ; $(info $(M) Stopping dev tools...) @  ## Stopping dev tools
		docker stop $(PACKAGE)_builder

build-docker: dev-start ; $(info $(M) Building sources within Docker...) @  ## Build the sources inside Denver
		docker exec $(PACKAGE)_builder bash -c "cd /go/src/$(PACKAGE); make lint test build-local; chown -R $(UID):$(GID) /go/src/$(PACKAGE)"

build-local: ; $(info $(M) Building sources...) @  ## Build the sources for CI
		$Q cd ./src && \
			for dist in $(DIST); do \
				GOOS=$$dist GOARCH=amd64 $(GO) build \
					-tags release \
					-ldflags '-X $(PACKAGE)/cmd.Version=$(VERSION) -X $(PACKAGE)/cmd.UpdatePath=$(S3PATH) -X $(PACKAGE)/cmd.WorkingDirectory=' \
					-o ../bin/$(PACKAGE)-$$dist ; \
			done ;

pack: ; $(info $(M) Packing releases...) @  ## Packing the releases
		$Q rm -rf ./releases
		$Q mkdir -p ./releases/$(DIST)/$(VERSION)/$(PACKAGE)/{conf,tools}
		$Q cd $(BASE) && \
			for dist in $(DIST); do \
				cp bin/$(PACKAGE)-$$dist releases/$$dist/$(VERSION)/$(PACKAGE)/$(PACKAGE) ; \
				cp conf/config.yml.dist releases/$$dist/$(VERSION)/$(PACKAGE)/conf/config.yml.dist ; \
				cp tools/alacritty-$$dist-* releases/$$dist/$(VERSION)/$(PACKAGE)/tools/ ; \
				cp tools/alacritty.yml releases/$$dist/$(VERSION)/$(PACKAGE)/tools/ ; \
			done
		$Q mv releases/windows/$(VERSION)/$(PACKAGE)/$(PACKAGE) releases/windows/$(VERSION)/$(PACKAGE)/$(PACKAGE).exe
		$Q cp tools/winpty-agent.exe releases/windows/$(VERSION)/$(PACKAGE)/tools/
		$Q cp tools/iterm2.sh releases/darwin/$(VERSION)/$(PACKAGE)/tools/
		$Q cd ./releases/linux && echo "{ \"filesize\": \"$$(du -s $(VERSION)/$(PACKAGE) | cut -f 1)\", \"date\": \"$(DATE)\", \"release\": \"$(VERSION)\", \"url\": \"$(S3PATH)/linux/$(PACKAGE)_$(VERSION)_Linux_amd64.tar.bz2\" }" > manifest.json
		$Q cd ./releases/linux/$(VERSION) && tar -I lbzip2 -cf ../$(PACKAGE)_$(VERSION)_Linux_amd64.tar.bz2 $(PACKAGE)
		$Q cd ./releases/darwin && echo "{ \"filesize\": \"$$(du -s $(VERSION)/$(PACKAGE) | cut -f 1)\", \"date\": \"$(DATE)\", \"release\": \"$(VERSION)\", \"url\": \"$(S3PATH)/darwin/$(PACKAGE)_$(VERSION)_Darwin_amd64.zip\" }" > manifest.json
		$Q cd ./releases/darwin/$(VERSION) && zip -rq ../$(PACKAGE)_$(VERSION)_Darwin_amd64.zip $(PACKAGE)
		$Q cd ./releases/windows && echo "{ \"filesize\": \"$$(du -s $(VERSION)/$(PACKAGE) | cut -f 1)\", \"date\": \"$(DATE)\", \"release\": \"$(VERSION)\", \"url\": \"$(S3PATH)/windows/$(PACKAGE)_$(VERSION)_Windows_amd64.zip\" }" > manifest.json
		$Q cd ./releases/windows/$(VERSION) && zip -rq ../$(PACKAGE)_$(VERSION)_Windows_amd64.zip $(PACKAGE)
		$Q for dist in $(DIST); do rm -rf ./releases/$$dist/$(VERSION) ; done

push-release-to-s3: ; $(info $(M) Push release to S3) @  ## Push release to S3
		$Q aws s3 sync --acl public-read ./releases s3://$(S3BUCKET)

clean: ; $(info $(M) Removing useless data...) @  ## Cleanup the project folder
		$Q -cd ./src && $(GO) clean

mrproper: clean ; $(info $(M) Remove useless data and binaries...) @  ## Clean everything and free resources
		$Q -rm -f  ./bin/*
		$Q -rm -rf ./store
		$Q -rm -rf .ssh
		$Q -rm -rf ./releases

########################################################
### External tools                                   ###
########################################################

install-tools: ; $(info $(M) Installing all tools...) @
	$Q go get -u golang.org/x/lint/golint
	$Q go get -u golang.org/x/tools/cmd/goimports
	$Q go get -u github.com/GeertJohan/fgt
	$Q go get -u github.com/kisielk/errcheck
	$Q cd ./src && go mod download
