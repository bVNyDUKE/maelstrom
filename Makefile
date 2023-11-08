SRCS := $(shell find ./app -name '*.go')
VOL := /home/kpanda/mlstrm/app:/app/test
w := echo

gobuild: $(SRCS)
	go build cmd

build: gobuild
	docker build . -t mlstrm

run: build
	docker run -it -v $(VOL) mlstrm

test: build
	docker run -it --rm -v $(VOL) mlstrm ./maelstrom test -w $(w) --bin test/mlstrm --nodes n1 --time-limit 10 --log-stderr
