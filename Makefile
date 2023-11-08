vol := /home/kpanda/mlstrm/app:/app/test
w := echo

gobuild:
	go build cmd

build:
	docker build . -t mlstrm

run: build
	docker run -it -v $(vol) mlstrm

test: build
	docker run -it -v $(vol) mlstrm ./maelstrom test -w $(w) --bin test/mlstrm --nodes n1 --time-limit 10 --log-stderr
