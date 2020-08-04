bin_name=stream_test_generate_data
pid_file=$(bin_name).pid

.IGNORE: stop

all: build start

clean:
	rm *.log
build:
	go build -o $(bin_name)

start:
	./$(bin_name) &

stop:
	@pkill -F $(pid_file) 2> /dev/null
