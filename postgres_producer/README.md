# Postgres Producer

This program will listen to channels on the postgres server and submit those messages to Kafka. the `stream_config.yaml` contains the connection information for both Postgres & Kafka as well as informatin about connecting Postgres notification channels to Kafka topics.
