package main

import (
	wapc "github.com/wapc/wapc-guest-tinygo"
)

func main() {
	wapc.RegisterFunctions(wapc.Functions{"savetoipfs": savetoipfs})
}

// rewrite returns a new URI if necessary.
func savetoipfs(request []byte) ([]byte, error) {
	res, err := wapc.HostCall("ipfs", "binding", "add", request)

	return res, err
}
