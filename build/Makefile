include .env
GIT_VERSION ?= $(shell git describe --tags --dirty --always)
GIT_COMMIT  ?= $(shell git rev-parse --short HEAD)
GIT_BRANCH  ?= $(shell git rev-parse --abbrev-ref HEAD)
PKG_VERSION =  $(shell go list ./version)

LDFLAGS += -X "${PKG_VERSION}.Project=${PROJECT}"
LDFLAGS += -X "${PKG_VERSION}.Version=${GIT_VERSION}"
LDFLAGS += -X "${PKG_VERSION}.GitCommit=${GIT_COMMIT}"
LDFLAGS += -X "${PKG_VERSION}.GitBranch=${GIT_BRANCH}"
LDFLAGS += -X "${PKG_VERSION}.BuildTS=$(shell date '+%Y-%m-%d %H:%M:%S')"

BUILD ?= dev

RACE_FLAG =
ifeq ("$(WITH_RACE)", "1")
        RACE_FLAG = -race
endif

COMMANDS ?= $(foreach DIR, $(dir $(wildcard cmd/*/.)), $(shell basename $(DIR)))
.PHONY: build
build: dep
	@mkdir -p bin
	for COMMAND in ${COMMANDS} ; do \
		go build -trimpath -v -o bin/$${COMMAND} -ldflags '$(LDFLAGS)' cmd/$${COMMAND}/*.go ; \
	done

.PHONY: dep
dep:
	go mod tidy -v


.PHONY: docker-build
docker-build:
	docker system prune --force
	docker build --progress=plain --rm -t ${DOCKER_TAG} \
		--build-arg PROJECT=${PROJECT} \
		--build-arg GIT_BRANCH=${GIT_BRANCH} \
		--build-arg BUILD=${BUILD} .
	docker tag ${DOCKER_TAG} ${DOCKER_TAG}:${GIT_VERSION}
	docker save ${DOCKER_TAG} | gzip > ${PROJECT}.${BUILD}.tgz

.PHONY: up
up:
	docker-compose -f docker-compose-db.yml -f docker-compose.yml up -d

.PHONY: down
down:
	docker-compose -f docker-compose-db.yml -f docker-compose.yml down

