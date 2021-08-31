.PHONY: default
all: default

app=$(notdir $(shell pwd))
goVersion := $(shell go version)
# echo ${goVersion#go version }
# strip prefix "go version " from output "go version go1.16.7 darwin/amd64"
goVersion2 := $(subst go version ,,$(goVersion))
buildTime := $(shell date '+%Y-%m-%d %H:%M:%S')
# https://git-scm.com/docs/git-rev-list#Documentation/git-rev-list.txt-emaIem
gitCommit := $(shell git rev-list --oneline --format=format:'%h@%aI' --max-count=1 `git rev-parse HEAD` | tail -1)
#gitCommit := $(shell git rev-list -1 HEAD)
# https://stackoverflow.com/a/47510909
pkg := github.com/bingoohuang/gg/pkg/v
#static := -static
# https://ms2008.github.io/2018/10/08/golang-build-version/
flags = "-extldflags $(static) -s -w -X '$(pkg).buildTime=$(buildTime)' -X $(pkg).appVersion=1.0.4 -X $(pkg).gitCommit=$(gitCommit) -X '$(pkg).goVersion=$(goVersion2)'"

default: install

tool:
	go get github.com/securego/gosec/cmd/gosec

sec:
	@gosec ./...
	@echo "[OK] Go security check was completed!"

init:
	export GOPROXY=https://goproxy.cn

lint:
	#golangci-lint run --enable-all
	golangci-lint run ./...

fmt:
	gofumports -w .
	gofumpt -w .
	gofmt -s -w .
	go mod tidy
	go fmt ./...
	revive .
	goimports -w .

install: init
	go install -ldflags="$(flags)" ./...
	ls -lh $$(which ${app})

linux: init
	GOOS=linux GOARCH=amd64 go install -ldflags="$(flags)" ./...
	upx ~/go/bin/linux_amd64/${app}

upx:
	ls -lh $$(which ${app})
	upx $$(which ${app})
	ls -lh $$(which ${app})
	ls -lh ~/go/bin/linux_amd64/${app}
	upx ~/go/bin/linux_amd64/${app}
	ls -lh ~/go/bin/linux_amd64/${app}

test: init
	#go test -v ./...
	go test -v -race ./...

bench: init
	#go test -bench . ./...
	go test -tags bench -benchmem -bench . ./...

clean:
	rm coverage.out

cover:
	go test -v -race -coverpkg=./... -coverprofile=coverage.out ./...

coverview:
	go tool cover -html=coverage.out

# https://hub.docker.com/_/golang
# docker run --rm -v "$PWD":/usr/src/myapp -v "$HOME/dockergo":/go -w /usr/src/myapp golang make docker
# docker run --rm -it -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang bash
# 静态连接 glibc
docker:
	mkdir -f ~/dockergo
	docker run --rm -v "$$PWD":/usr/src/myapp -v "$$HOME/dockergo":/go -w /usr/src/myapp golang make dockerinstall
	#upx ~/dockergo/bin/${app}
	gzip -f ~/dockergo/bin/${app}

dockerinstall:
	go install -v -x -a -ldflags '-extldflags "-static"' ./...
