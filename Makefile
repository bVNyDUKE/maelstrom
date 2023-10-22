build:
	docker build . -t mlstrm

run:
	docker run -it -v /home/kpanda/mlstrm:/app mlstrm

test:
	./maelstrom test -w echo --bin mlstrm --nodes n1 --time-limit 10 --log-stderr
