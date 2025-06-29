module github.com/xjhc/alignment/server

go 1.21

require (
	github.com/google/uuid v1.4.0
	github.com/gorilla/websocket v1.5.1
	github.com/redis/go-redis/v9 v9.3.0
	github.com/xjhc/alignment/core v0.0.0
)

replace github.com/xjhc/alignment/core => ../core

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	golang.org/x/net v0.17.0 // indirect
)
