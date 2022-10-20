package main

import (
	"strings"

	wapc "github.com/wapc/wapc-guest-tinygo"
)

func main() {
	wapc.RegisterFunctions(wapc.Functions{"upperCase": upperCase})
}

// rewrite returns a new URI if necessary.
func upperCase(request []byte) ([]byte, error) {
	return []byte(strings.ToUpper(string(request))), nil
}
