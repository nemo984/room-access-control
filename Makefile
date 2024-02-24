build:
	docker build -t rac .

run:
	docker run -p 8080:8080 --env-file .env rac
