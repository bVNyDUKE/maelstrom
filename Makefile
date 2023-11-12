APP_NAME := mlstrm
DOCKER_IMAGE_NAME := mlstrm-img
DOCKER_CONTAINER_NAME := mlstrm-container
APP_DIR := ./app
GO_SRCS := $(wildcard $(APP_DIR)/*.go)

w := echo

$(APP_NAME): $(GO_SRCS)
	cd $(APP_DIR) && go build -o ../$(APP_NAME)

build: $(APP_NAME) ./Dockerfile
	docker build . -t $(DOCKER_IMAGE_NAME)

run: build
	docker run --name $(DOCKER_CONTAINER_NAME) -it $(DOCKER_IMAGE_NAME)

test: build
	docker run --name $(DOCKER_CONTAINER_NAME) -it --rm $(DOCKER_IMAGE_NAME) ./maelstrom test -w $(w) --bin mlstrm --nodes n1 --time-limit 10 --log-stderr

generate: build
	docker run --name $(DOCKER_CONTAINER_NAME) -it --rm $(DOCKER_IMAGE_NAME) ./maelstrom test -w unique-ids --bin mlstrm --time-limit 30 --rate 1000 --node-count 3 --availability total --nemesis partition

clean: 
	rm -f $(APP_NAME)
	docker rm $(DOCKER_CONTAINER_NAME) 2>/dev/null || true
	docker rmi $(DOCKER_IMAGE_NAME) 2>/dev/null || true

.PHONY: clean
