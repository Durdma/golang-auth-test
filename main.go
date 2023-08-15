package main

import (
	"test-auth/auth"
	"test-auth/server"
)

func main() {
	auth.GetHash()
	server.Init()
}
