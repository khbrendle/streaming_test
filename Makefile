DOCKER_PG=stream-test-pg
DOCKER_ZOOKEEPER=stream-test-zookeeper
DOCKER_KAFKA=stream-test-kafka


docker-delete-all: postgres-docker-delete zookeeper-docker-delete kafka-docker-delete
	@echo 'all dockers destroyed'

postgres-docker-init:
	docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=webapp --name=$(DOCKER_PG) postgres
postgres-docker-start:
	docker start $(DOCKER_PG)
postgres-docker-stop:
	docker stop $(DOCKER_PG)
postgres-docker-delete:
	-docker rm $(DOCKER_PG)
postgres-db-init:
	psql -h localhost -p 5432 -U postgres -f db/create_tables.sql


zookeeper-docker-init:
	docker run -d \
		--name $(DOCKER_ZOOKEEPER) \
		-p 2181:2181 \
		--restart=always \
		zk-alpine:0.1
zookeeper-docker-start:
	docker start $(DOCKER_ZOOKEEPER)
zookeeper-docker-stop:
	docker stop $(DOCKER_ZOOKEEPER)
zookeeper-docker-delete:
	-docker rm $(DOCKER_ZOOKEEPER)


kafka-docker-init:
	docker run -d \
		--name $(DOCKER_KAFKA) \
		-p 9092:9092 \
		--link=$(DOCKER_ZOOKEEPER) \
		--restart=always \
		kafka-alpine:0.2
kafka-docker-start:
	docker start $(DOCKER_KAFKA)
kafka-docker-stop:
	docker stop $(DOCKER_KAFKA)
kafka-docker-delete:
	-docker rm $(DOCKER_KAFKA)
