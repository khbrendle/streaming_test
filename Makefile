DOCKER_PG=stream-test-pg
DOCKER_ZOOKEEPER=stream-test-zookeeper
DOCKER_KAFKA=stream-test-kafka

ui-start:
	$(MAKE) -C ui run

stream-start:
	@echo 'setting up database'
	$(MAKE) -C db all
	@echo 'starting Postgres listener'
	$(MAKE) -C postgres_producer all
	@echo 'starting backend web server'
	$(MAKE) -C stream_server all
	@echo 'generating data'
	$(MAKE) -C generate_data all
stream-stop:
	@echo 'stopping initializing db'
	@$(MAKE) -C db stop
	@echo 'stopping generating data'
	@$(MAKE) -C generate_data stop
	@echo 'stopping Postgres listener'
	@$(MAKE) -C postgres_producer stop
	@echo 'stopping backend web server'
	@$(MAKE) -C stream_server stop


docekr-init-all: postgres-docker-init zookeeper-docker-init kafka-docker-init
	@echo 'initializing all dockers'
docekr-start-all: postgres-docker-start zookeeper-docker-start kafka-docker-start
	@echo 'starting all dockers'
docekr-stop-all: postgres-docker-stop zookeeper-docker-stop kafka-docker-stop
	@echo 'stopping all dockers'
docker-delete-all: postgres-docker-delete zookeeper-docker-delete kafka-docker-delete
	@echo 'deleting all dockers'


postgres-docker-init:
	docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=webapp --name=$(DOCKER_PG) postgres
postgres-docker-start:
	docker start $(DOCKER_PG)
postgres-docker-stop:
	docker stop $(DOCKER_PG)
postgres-docker-delete:
	-docker rm $(DOCKER_PG)


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
