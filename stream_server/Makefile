bin_name=stream_test_stream_server
pid_file=$(bin_name).pid

.IGNORE: stop

all: build start

clean:
	rm *.log
build:
	@echo 'building web server'
	go build -o $(bin_name)

start:
	@echo 'starting web server'
	./$(bin_name) &

stop:
	@echo 'stopping web server'
	@pkill -F $(pid_file) 2> /dev/null
