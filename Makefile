d := docker
vol := /home/kpanda/mlstrm/app:/app/test
r := $(d) run -it -v $(vol) mlstrm
w := echo

gobuild:
	go build cmd

build:
	$(d) build ./build/Dockerfile -t mlstrm

run: build
	$(r)

test: build
	$(r) ./maelstrom test -w $(w) --bin test/mlstrm --nodes n1 --time-limit 10 --log-stderr
