package main

import (
	"strings"

	wapc "github.com/wapc/wapc-guest-tinygo"
)

func main() {
	wapc.RegisterFunctions(wapc.Functions{"actorCall": upperCase})
}

// rewrite returns a new URI if necessary.
func upperCase(request []byte) ([]byte, error) {
	return wapc.HostCall("bafybeig3akf4jm5xerhyh5fm6skzsyhrhdug63fl6eyuov7ggpv7yzsjdq", "actor", "savetoipfs", []byte(strings.ToUpper(string(request))))
}
