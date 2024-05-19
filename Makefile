all:
	cd client && go build;
	cd server && go build;

udpProxy: udpProxy/client/main.go udpProxy/server/main.go
	cd udpProxy/client && go build;
	cd udpProxy/server && go build;

delete:
	rm server/server;
	rm server/tls_key.log;
	rm client/client;
	rm udpProxy/server/server;
	rm udpProxy/client/client;