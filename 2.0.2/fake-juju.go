package main

import (
	"os"
	"./service"
)

func main() {
	code := 0
	service.RunFakeJuju()
	os.Exit(code)
}
