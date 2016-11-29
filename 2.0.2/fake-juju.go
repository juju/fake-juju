package main

import (
	"./service"
)

func main() {
	code := 0
	service.RunFakeJuju()
	os.Exit(code)
}
