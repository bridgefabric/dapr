package w3s

import (
	"context"
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"os"

	"github.com/ipfs/go-cid"
	"github.com/web3-storage/go-w3s-client"
)

const token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJkaWQ6ZXRocjoweGFmNDQ2OUU3YjgwODU3Nzk0ODk2YzAzNEIzOUI0OTFBN0QzRGViMDEiLCJpc3MiOiJ3ZWIzLXN0b3JhZ2UiLCJpYXQiOjE2NjMxNDAwNDUzOTIsIm5hbWUiOiJ6YyJ9.KaMnsAxHxdwcSrAFrgPqORbhiWmxgFsfrck3cjlqHZY"

var client w3s.Client

func init() {
	err := Init(token)
	if err != nil {
		panic(err)
	}
}

func Init(token string) (err error) {
	client, err = w3s.NewClient(w3s.WithToken(token))
	if err != nil {
		return
	}
	return
}

func PutJson(ctx context.Context, data interface{}) (cid string, err error) {
	jsonByes, err := json.Marshal(data)
	if err != nil {
		return
	}
	return Put(ctx, jsonByes)
}

func Put(ctx context.Context, data []byte) (cid string, err error) {
	panic("todo")
}

func PutFileByPath(ctx context.Context, path string) (cid string, err error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	return PutFile(ctx, file)
}

func PutFile(ctx context.Context, file *os.File) (cid string, err error) {
	cidObj, err := client.Put(ctx, file)
	if err != nil {
		return
	}
	return cidObj.String(), nil
}

func Get(ctx context.Context, cidStr string) (data []byte, err error) {
	cidObj, err := cid.Decode(cidStr)
	if err != nil {
		return nil, err
	}
	res, err := client.Get(ctx, cidObj)
	if err != nil {
		return nil, err
	}
	_, fsys, err := res.Files()
	if err != nil {
		return nil, err
	}
	err = fs.WalkDir(fsys, "/", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			file, err := fsys.Open(path)
			if err != nil {
				return err
			}
			data, err = ioutil.ReadAll(file)
			if err != nil {
				return err
			}
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return
}
