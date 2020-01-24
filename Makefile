build: build-rb build-go

build-rb:
	docker build -f Dockerfile-rb -t sitemap-test-rb .

build-go:
	docker build -f Dockerfile-go -t sitemap-test-go .

network = sitemap-test
docker_args = -t --rm -e AWS_ACCESS_KEY_ID=minioadmin -e AWS_SECRET_ACCESS_KEY=minioadmin --network $(network) -e AWS_ENDPOINT=http://minio:9000

run-go:
	docker run $(docker_args) sitemap-test-go

run-rb:
	docker run $(docker_args) sitemap-test-rb

run-minio:
	docker network create $(network)
	docker run -d --rm --name minio -p 9000:9000 --network $(network) minio/minio server miniodata

clean:
	docker stop minio
	docker network rm $(network)
