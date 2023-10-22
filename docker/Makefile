test:
	./maelstrom test -w echo --bin /maelstrom/golang/mlstrm --nodes n1 --time-limit 10 --log-stderr

build:
	docker build . -t mlstrm

run:
	docker run -it -v /home/kpanda/mlstrm/golang:/maelstrom/golang mlstrm
