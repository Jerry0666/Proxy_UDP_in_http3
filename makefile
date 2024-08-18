
all:
	cd client && go build;
	cd server && go build;
	cd control && go build;

proxy: client/main.go server/main.go
	cd client && go build;
	cd server && go build;


deleteProxy:
	rm client/client;
	rm server/server;
	rm control/control;
	rm -f server/tls_key.log;

delete:
	rm client/client;
	rm server/server;
	rm -f server/tls_key.log;
	rm control/control;

server:
	iperf -s -u -p 8000 -i 1 -l 1300

client:
	iperf -c 201.0.0.1 -u -p 8000 -l 1300 -i 1 -b 30M -t 10 -R