bin_name=stream_test_postgres_producer
pid_file=$(bin_name).pid

.IGNORE: stop

all: build start

clean:
	rm *.log
build:
	go build -o $(bin_name)
start:
	./$(bin_name) "stream_config.yaml" &
stop:
	@pkill -F $(pid_file) 2> /dev/null
