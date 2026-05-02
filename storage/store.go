package storage

import (
	"fmt"
	"os"
)

const (
	BasePath = "./data"
)

func SaveChunk(chunkId string, data []byte) (int, error) {
	path := fmt.Sprintf("%s/%s", BasePath, chunkId)

	file, err := os.Create(path)

	if err != nil {
		return 0,err
	}

	defer file.Close()

	n, err := file.Write(data)
	return n, nil
}

func ReadChunk(chunkId string) ([]byte, error) {
	path := fmt.Sprintf("%s/%s", BasePath, chunkId)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return data, nil
}
