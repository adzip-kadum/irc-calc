include build/Makefile

.PHONY: freebsd-build
freebsd-build:
	GOOS=freebsd make

.PHONY: docker-build
docker-build:
	docker build --progress=plain --rm -t ${DOCKER_TAG} --build-arg GIT_BRANCH=${GIT_BRANCH} .
	docker tag ${DOCKER_TAG} ${DOCKER_TAG}:${GIT_VERSION}
	docker save ${DOCKER_TAG} | gzip > irc.latest.tgz
