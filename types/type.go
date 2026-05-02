package types

type ChunkMeta struct {
	ChunkId string `json:"chunk_id"`
	Index   int    `json:"index"`
}

type FileData struct {
	FileId   string `json:"file_id"`
	FileName string `json:"filename"`
}

type FileMeta struct {
	File       FileData    `json:"file"`
	ChunkCount int         `json:"chunk_count"`
	Chunks     []ChunkMeta `json:"chunks"`
}
