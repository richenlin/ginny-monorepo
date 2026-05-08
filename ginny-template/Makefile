# fetches this repo into $GOPATH
# go install github.com/envoyproxy/protoc-gen-validate@latest
# go install github.com/golang/protobuf/protoc-gen-go@latest

# 发布时根据情况修改
APP = APP_NAME
CONF ?= config.yml
# amd64 arm64
ARCH ?= amd64
PROTO_IMG = ginny/protoc:latest

REPO         = docker.io/username
IMG_REPO     := $(REPO)/$(APP)
#-------------------------------------	
.PHONY: run
run: tidy proto wire
	go run ./cmd/ -f configs/$(CONF)  & \
#-------------------------------------	
.PHONY: tidy
tidy:
	go mod tidy
#-------------------------------------	
.PHONY: wire
wire: 
	wire ./...
#-------------------------------------	
.PHONY: test
test: tidy mock
	go test -v ./internal/... -f `pwd`/configs/$(CONF) -covermode=count -coverprofile=dist/cover-$(APP).out
#-------------------------------------	
.PHONY: build
build: tidy wire
	GOOS=linux GOARCH=$(ARCH) go build -o "dist/$(APP)-$(ARCH).bin" ./cmd/;
#-------------------------------------	
.PHONY: img
img: build
	docker buildx build -f ./deploy/Dockerfile -t $(IMG_REPO):$(IMG_VERSION) .
#-------------------------------------	
.PHONY: cover
cover: test
	go tool cover -html=dist/cover-$(APP).out
#-------------------------------------	
.PHONY: mock
mock:
	mockery --all
#-------------------------------------	
.PHONY: lint
lint:
	golint ./...
#-------------------------------------	
.PHONY: protoc
protoc:
	docker run --rm -v $(shell pwd):/build/go -v $(shell pwd):/build/proto -v $(shell pwd)/doc:/build/openapi ${PROTO_IMG}
# protoc -I api/proto --go_out="plugins=grpc:api/proto" --validate_out="lang=go:api/proto" ./api/proto/*.proto
#-------------------------------------	
# .PHONY: docker
# docker-compose: build dash rules
# 	docker-compose -f deploy/docker-compose.yml up --build -d
all: lint cover test
