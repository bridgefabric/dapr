package w3s

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpload(t *testing.T) {
	file, err := os.Open("../examples/uppercase/example.wasm")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	cid, err := PutFile(context.Background(), file)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(cid))

	file2, err := os.Open("../examples/uppercase/example.wasm")
	if err != nil {
		t.Fatal(err)
	}
	defer file2.Close()
	bytes, err := io.ReadAll(file2)
	if err != nil {
		t.Fatal(err)
	}

	res, err := Get(context.Background(), cid)
	if err != nil {
		t.Fatal(err)
	}
	for i := len(res) - 1; i >= 0; i-- {
		if len(res) > i && res[i] == bytes[i] {
			continue
		}
		fmt.Println("not equeal", i)
	}
	assert.Equal(t, res, bytes)
}
