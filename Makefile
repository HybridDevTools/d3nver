ifndef VERBOSE
.SILENT:
endif

PACKAGE        := denver
DATE           := $(shell date +%s)
VERSION        := $(shell git --no-pager log --pretty=format:'%h' -n 1)
SHELL          := bash

BUILDER_IMG    := golang:stretch
BUILDER_STATUS := $(shell docker inspect -f '{{.State.Running}}' denver_builder 2> /dev/null)

GO             := go
BASE           := $(shell pwd)

DIST           := {linux,darwin,windows}

S3BUCKET       := s3.d3nver.io/app
RELEASE        := v$(VERSION)-$(DATE)
PROJECT_ID     := 181
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

build-denver: dev-start ; $(info $(M) Building sources within Docker...) @  ## Build the sources inside Denver
		docker exec $(PACKAGE)_builder bash -c "cd /go/src/$(PACKAGE); make lint test build-ci; chown -R $(UID):$(GID) /go/src/$(PACKAGE)"

build-ci: ; $(info $(M) Building sources...) @  ## Build the sources for CI
		$Q cd ./src && \
			for dist in $(DIST); do \
				GOOS=$$dist GOARCH=amd64 $(GO) build \
					-tags release \
					-ldflags '-X $(PACKAGE)/cmd.Version=v$(VERSION) -X $(PACKAGE)/cmd.BuildTs=$(DATE) -X $(PACKAGE)/cmd.WorkingDirectory=' \
					-o ../bin/$(PACKAGE)-$$dist ; \
			done ; \
			echo "$(DATE)" > ../bin/build-ts.txt ; \
			echo "$(RELEASE)" > ../bin/build-release.txt ; \

pack: ; $(info $(M) Packing releases...) @  ## Packing the releases
		$(eval DATE    := $(shell cat ./bin/build-ts.txt))
		$(eval RELEASE := $(shell cat ./bin/build-release.txt))
		$Q rm -rf ./releases
		$Q mkdir -p ./releases/$(DIST)/$(DATE)/$(PACKAGE)/{conf,tools}
		$Q cd $(BASE) && \
			for dist in $(DIST); do \
				cp bin/$(PACKAGE)-$$dist releases/$$dist/$(DATE)/$(PACKAGE)/$(PACKAGE) ; \
				cp conf/config.yml.dist releases/$$dist/$(DATE)/$(PACKAGE)/conf/config.yml.dist ; \
				cp tools/alacritty-$$dist-* releases/$$dist/$(DATE)/$(PACKAGE)/tools/ ; \
				cp tools/alacritty.yml releases/$$dist/$(DATE)/$(PACKAGE)/tools/ ; \
			done
		$Q mv releases/windows/$(DATE)/$(PACKAGE)/$(PACKAGE) releases/windows/$(DATE)/$(PACKAGE)/$(PACKAGE).exe
		$Q cp tools/winpty-agent.exe releases/windows/$(DATE)/$(PACKAGE)/tools/
		$Q cp tools/iterm2.sh releases/darwin/$(DATE)/$(PACKAGE)/tools/
		$Q cd ./releases/linux && echo "{ \"filesize\": \"$$(du -s $(DATE)/$(PACKAGE) | cut -f 1)\", \"date\": \"$(DATE)\", \"release\": \"$(RELEASE)\", \"url\": \"$(S3PATH)/linux/$(DATE)/$(PACKAGE)-linux-$(RELEASE).tar.bz2\" }" > manifest.json
		$Q cd ./releases/linux/$(DATE) && tar -I lbzip2 -cf ./$(PACKAGE)-linux-$(RELEASE).tar.bz2 $(PACKAGE)
		$Q cd ./releases/darwin && echo "{ \"filesize\": \"$$(du -s $(DATE)/$(PACKAGE) | cut -f 1)\", \"date\": \"$(DATE)\", \"release\": \"$(RELEASE)\", \"url\": \"$(S3PATH)/darwin/$(DATE)/$(PACKAGE)-darwin-$(RELEASE).zip\" }" > manifest.json
		$Q cd ./releases/darwin/$(DATE) && zip -rq ./$(PACKAGE)-darwin-$(RELEASE).zip $(PACKAGE)
		$Q cd ./releases/windows && echo "{ \"filesize\": \"$$(du -s $(DATE)/$(PACKAGE) | cut -f 1)\", \"date\": \"$(DATE)\", \"release\": \"$(RELEASE)\", \"url\": \"$(S3PATH)/windows/$(DATE)/$(PACKAGE)-windows-$(RELEASE).zip\" }" > manifest.json
		$Q cd ./releases/windows/$(DATE) && zip -rq ./$(PACKAGE)-windows-$(RELEASE).zip $(PACKAGE)
		$Q for dist in $(DIST); do rm -rf ./releases/$$dist/$(DATE)/$(PACKAGE) ; done

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
