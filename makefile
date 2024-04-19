
all:
	cd client && go build;
	cd server && go build;
	cd application/client && go build;
	cd application/server && go build;

delete:
	rm client/client;
	rm server/server;
	rm application/server/server;
	rm application/client/client;
	rm -f server/tls_key.log;