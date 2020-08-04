# Streaming Test

This repository sets up a simple data streaming service. Modeled after a sales type of data, customers & orders are generated which populate individual records and trigger inserts to fact tables for topics of Customer & Order. Inserts into fact tables create Postgres notifications which are listened for and sent to Kafka.

The web server allows clients to select from available data streams and watch customers be created and orders be placed.

### Setting up

To set up this demo application we will need to run a few containers

- Postgres
  I use a postgres docker container to keep things simple which can be spun up from the project Makefile
  `make postgres-docker-init`
- Zookeeper
  Zookeeper is the cluster coordinator, it needs to be started before Kafka. The following command will start Zookeeper
  `make zookeeper-docker-init`

  *zookeeper conatiner image probably needs to be created see [zookeeper docker](https://github.com/khbrendle/docker/tree/master/zookeeper)*
- Kafka
  Kafka is our message broker, relies on Zookeeper and proivdes the data stream.
  `make kafka-docker-init`

  This Kafka container does have an issues in that is tries to reach itself for comminucation and will try to access itself by the Docker container ID as the host. This container ID does not resolve, the simple solution is to add a record in `/etc/hosts` file `127.0.0.1 containerID`

  *kafka conatiner image probably needs to be created see [kafka docker](https://github.com/khbrendle/docker/tree/master/kafka)*

### Running
With the containers set up we get get it all running.

main directory makefile has 2 commands:
  - `make ui-start` will run the node UI server with hot-reload in dev mode

  - `make stream-start` will initialize the database tables, start the postgres listener/producer, start generating data, start the API server

  - `make stream-stop` will stop the programs running realted to generating & streaming the data

  - there are also a few make directives for managing all the docker containers
    - `make docker-init-all`
    - `make docker-start-all`
    - `make docker-stop-all`
    - `make docker-delete-all`
