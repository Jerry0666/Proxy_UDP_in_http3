
all:
	cd client && go build;
	cd server && go build;
	cd application/client && go build;
	cd application/server && go build;

proxy: client/main.go server/main.go
	cd client && go build;
	cd server && go build;

deleteProxy:
	rm client/client;
	rm server/server;
	rm -f server/tls_key.log;
	
delete:
	rm client/client;
	rm server/server;
	rm -f server/tls_key.log;
	rm 1480/client/client;
	rm 1480/server/server;
	rm bandwidth_test/client/client;
	rm bandwidth_test/server/server;