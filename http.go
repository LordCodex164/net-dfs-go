package main

// import (
// 	"fmt"
// 	"log"
// 	"net/http"

// 	"github.com/LordCodex164/net-protocol/redis"
// 	"github.com/LordCodex164/net-protocol/storage"
// 	"github.com/LordCodex164/net-protocol/types"
// 	"github.com/LordCodex164/net-protocol/utils"
// )

// var redisClient = redis.NewRedisClient()

// // split the chunks of a file and save the meta
// func uploadHandler(w http.ResponseWriter, r *http.Request) {
// 	file, header, err := r.FormFile("file")

// 	if err != nil {
// 		http.Error(w, err.Error(), 400)
// 	}

// 	defer file.Close()

// 	fmt.Println(header.Filename, "filename")

// 	fileId := GenerateID(header.Filename)

// 	_, chunks, err := utils.SplitToChunks(file)

// 	if err != nil {
// 		http.Error(w, err.Error(), 500)
// 		log.Fatal(err)
// 	}

// 	var chunkM []types.ChunkMeta

// 	for i, chunk := range chunks {
// 		fmt.Println("chunk", chunk[0])
// 		chunkId := generateChunkID(fileId, i)
// 		_, err := storage.SaveChunk(chunkId, chunk)
// 		if err != nil {
// 			http.Error(w, err.Error(), 400)
// 			log.Fatal(err)
// 		}
// 		chunkMeta := &types.ChunkMeta{
// 			ChunkId: chunkId,
// 			Index:   i,
// 		}
// 		chunkM = append(chunkM, *chunkMeta)
// 	}

// 	meta := types.FileMeta{
// 		File: types.FileData{
// 			FileId:   fileId,
// 			FileName: header.Filename,
// 		},
// 		ChunkCount: len(chunks),
// 		Chunks:     chunkM,
// 	}

// 	err = redis.SaveMetadata(redisClient, meta, fileId)
// 	if err != nil {
// 		http.Error(w, err.Error(), 500)
// 		log.Fatal(err)
// 	}

// 	fmt.Printf("%+v", meta)
// 	w.Write([]byte("uploading file"))

// }

// func downloadHandler(w http.ResponseWriter, r *http.Request) {

// 	if r.Method != http.MethodGet {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	fileId := r.URL.Query().Get("file_id")

// 	metaD, err := redis.GetMetaData(redisClient, fileId)

// 	if err != nil {
// 		http.Error(w, err.Error(), 500)
// 		log.Fatal(err)
// 	}

// 	fmt.Printf("meta: %+v", metaD)

// 	for _, chunk := range metaD.Chunks {
// 		fileData, _ := storage.ReadChunk(chunk.ChunkId)
// 		_, _ = w.Write(fileData)
// 	}

// }
