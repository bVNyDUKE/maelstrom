APP_NAME := mlstrm
DOCKER_IMAGE_NAME := mlstrm-img
DOCKER_CONTAINER_NAME := mlstrm-container
APP_DIR := ./app
GO_SRCS := $(wildcard $(APP_DIR)/*.go)
EXEC := ./maelstrom-sh
TEST := docker run --name $(DOCKER_CONTAINER_NAME) -it --rm $(DOCKER_IMAGE_NAME) $(EXEC) test 

# ./maelstrom-sh test -w broadcast --bin mlstrm --time-limit 20 --rate 100 --latency 100 --node-count 25

$(APP_NAME): $(GO_SRCS)
	cd $(APP_DIR) && go build -o ../$(APP_NAME)

build: $(APP_NAME) ./Dockerfile
	docker build . -t $(DOCKER_IMAGE_NAME)

run: build
	docker run --name $(DOCKER_CONTAINER_NAME) -it $(DOCKER_IMAGE_NAME)

test: build
	$(TEST) -w echo --bin mlstrm --nodes n1 --time-limit 10 --log-stderr

generate: build
	$(TEST) -w unique-ids --bin mlstrm --time-limit 30 --rate 1000 --node-count 3 --availability total --nemesis partition

broadcast: build
	$(TEST) -w broadcast --bin mlstrm --time-limit 20 --rate 10 --node-count 5

broadcast-partition: build
	$(TEST) -w broadcast --bin mlstrm --time-limit 20 --rate 10 --node-count 5 --nemesis partition

broadcast-perf: build
	$(TEST) -w broadcast --bin mlstrm --time-limit 20 --rate 100 --latency 100 --node-count 25

clean: 
	rm -f $(APP_NAME)
	docker rm $(DOCKER_CONTAINER_NAME) 2>/dev/null || true
	docker rmi $(DOCKER_IMAGE_NAME) 2>/dev/null || true

.PHONY: clean
