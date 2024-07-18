
all:
	cd client && go build;
	cd server && go build;
	cd application/client && go build;
	cd application/server && go build;

proxy: client/main.go server/main.go
	cd client && go build;
	cd server && go build;
	cd control && go build;

deleteProxy:
	rm client/client;
	rm server/server;
	rm control/control;
	rm -f server/tls_key.log;
	
delete:
	rm client/client;
	rm server/server;
	rm -f server/tls_key.log;
	rm 1480/client/client;
	rm 1480/server/server;
	rm bandwidth_test/client/client;
	rm bandwidth_test/server/server;

server:
	iperf -s -u -p 8000 -i 1 -l 1300

client:
	iperf -c 201.0.0.1 -u -p 8000 -l 1300 -i 1 -b 10M -t 20