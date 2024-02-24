.DEFAULT_GOAL := r

build:
	docker build -t rac .

run:
	docker run --rm -p 8080:8080 --env-file .env rac

r:
	./run.sh