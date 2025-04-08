.PHONY: all
all: build docker-build lint

.PHONY: release
release: docker-push

define PROMPT
	@echo " "
	@echo "**********************************************************"
	@echo "*"
	@echo "*   $(1)"
	@echo "*"
	@echo "**********************************************************"
	@echo " "
endef

BINARY_NAME=rbd-exporter
BINARIES=\
	./out/linux-amd64/$(BINARY_NAME) \
	./out/linux-arm64/$(BINARY_NAME) \
	./out/darwin-amd64/$(BINARY_NAME)

TAG?=latest

./out/linux-amd64/% : GOOS=linux
./out/linux-amd64/% : GOARCH=amd64
./out/linux-arm64/% : GOOS=linux
./out/linux-arm64/% : GOARCH=arm64
./out/darwin-amd64/% : GOOS=darwin
./out/darwin-amd64/% : GOARCH=amd64

.PHONY: $(BINARIES)
$(BINARIES):
	$(call PROMPT,$@)
	GOARCH=$(GOARCH) GOOS=$(GOOS) CGO_ENABLED=0 go build -o $@ ./cmd/rbd-exporter

.PHONY: build
build: $(BINARIES)

.PHONY: test
test:
	$(call PROMPT,$@)
	go tool gocov test -v ./... | go tool gocov-html > coverage.html

.PHONY: lint
lint:
	$(call PROMPT,$@)
	golangci-lint run

.PHONY: docker-build
docker-build: ./out/linux-amd64/$(BINARY_NAME)
	$(call PROMPT,$@)
	docker build --platform linux/amd64 -t boyvinall/rbd-exporter:$(TAG) .

.PHONY: docker-push
docker-push: docker-build
	$(call PROMPT,$@)
	docker push boyvinall/rbd-exporter:$(TAG)