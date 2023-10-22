gobuild:
	go build cmd

build:
	docker build ./build/Dockerfile -t mlstrm

run:
	docker run -it -v /home/kpanda/mlstrm/app:/app/test mlstrm

echo:
	./maelstrom test -w echo --bin test/mlstrm --nodes n1 --time-limit 10 --log-stderr
