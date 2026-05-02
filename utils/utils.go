package utils

import "io"

func SplitToChunks(src io.Reader) (int, [][]byte, error) {

	var chunks [][]byte
	nw := 0

	for {
		buf := make([]byte, 10*1024)
		n, err := src.Read(buf)
		if n > 0 {
			chunks = append(chunks, buf[:n])
			nw += n
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, nil, err
		}
	}
	return nw, chunks, nil
}
