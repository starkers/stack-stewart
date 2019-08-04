.DEFAULT_GOAL := help


APP     := starkers/stack-stewart

GO111MODULES = on

# make build will create these files
SERVER_OUT := "server.bin"
AGENT_OUT := "agent.bin"


PKG := "github.com/starkers/stack-stewart"
SERVER_PKG_BUILD := "${PKG}/cmd/server"
AGENT_PKG_BUILD := "${PKG}/cmd/agent"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)

.PHONY: all api build_server build_agent

all: api build_server build_agent ## do all the things

dep: ## Get the dependencies
	go get -v -d ./...

build: build_agent build_server

build_server: dep api ## Build the binary file for server
	go build -i -v -o $(SERVER_OUT) $(SERVER_PKG_BUILD)

build_agent: dep api ## Build the binary file for agent
	go build -i -v -o $(AGENT_OUT) $(AGENT_PKG_BUILD)

clean: ## Remove previous builds
	rm $(SERVER_OUT) $(AGENT_OUT)

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


##########

# ensure you create an initial commit on your your git.. `git tag 0.0.1 ; git push origin 0.0.1`
export TAG     := $(shell git describe --tags)
export IMG     := "$(APP):$(TAG)"

# DOCKER TASKS
# Build the container
docker_build: ## Build the container
	docker build -t $(IMG) .

publish-all: login publish-version publish-latest

publish-latest: tag-latest login ## Publish the `latest` taged container to ECR
	@echo 'publish latest to $(DOCKER_REPO)'
	docker push $(APP):latest

publish-version: login ## Publish the `{version}` taged container to ECR
	@echo 'publish $(TAG)'
	docker push $(APP):$(TAG)

tag-latest: ## Generate container `{version}` tag
	@echo 'create tag latest'
	docker tag $(APP):$(TAG) $(APP):latest

# log into dockerhub
login:
	docker login -u $(DOCKER_USER) -p $(DOCKER_PASS)
