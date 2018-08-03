package files

import (
	"crypto/md5"
	"encoding/base64"
	"math"
	"os"
	"path/filepath"
	"strings"
)

const (
	minChunkSize     = 64000
	maxChunksPerFile = 128
)

// FileItems map of file items
type FileItems map[string]FileItem

// FileItem file info containing size, hash and chunk data and sub items if it's a dir.
type FileItem struct {
	Size      int64     `json:"size"`
	Hash      string    `json:"hash"`
	Directory bool      `json:"directory"`
	Chunks    []Chunk   `json:"chunks"`
	Items     FileItems `json:"items"`
}

// ChunkSlice slice of Chunks
type ChunkSlice []Chunk

// Chunk - info for a small segment of a file
type Chunk struct {
	Size int64  `json:"size"`
	Hash string `json:"hash"`
}

// Count - Return the total number of items recursively
func (fi FileItems) Count() int {
	result := 0
	for _, f := range fi {
		if !f.Directory {
			result++
		} else if f.Items != nil {
			result += f.Items.Count()
		}
	}
	return result
}

// Size - Return the total size in bytes of items recursively
func (fi FileItems) Size(other FileItems) int64 {
	var result int64
	for name, f := range fi {
		otherF, otherOk := other[name]
		if !f.Directory {
			if !otherOk {
				result += f.Size
				continue
			}
			for i, chunk := range f.Chunks {
				if len(otherF.Chunks) >= i {
					result += chunk.Size
					continue
				}
				if otherF.Chunks[i].Hash != chunk.Hash {
					result += chunk.Size
				}
			}
		} else if f.Items != nil {
			otherItems := otherF.Items
			if otherItems == nil {
				otherItems = make(FileItems)
			}
			result += f.Items.Size(otherItems)
		}
	}
	return result
}

// CalculateHash - calculate the total hash of a chunk slice
func (cs ChunkSlice) CalculateHash() string {
	hashes := make([]string, len(cs))
	for i, c := range cs {
		hashes[i] = c.Hash
	}
	hashData := md5.Sum([]byte(strings.Join(hashes, "_")))
	return base64.URLEncoding.EncodeToString(hashData[:])
}

// GetFileItems - return the file data for the given basepath
func GetFileItems(basepath string) (FileItems, error) {
	filePaths, err := filepath.Glob(filepath.Join(basepath, "*"))
	if err != nil {
		return map[string]FileItem{}, err
	}

	fileItems := make(map[string]FileItem)
	for _, fPath := range filePaths {
		f, err := os.Open(fPath)
		if err != nil {
			return fileItems, err
		}

		defer f.Close()
		fi, err := f.Stat()
		if err != nil {
			return fileItems, err
		}

		name := filepath.Base(fPath)
		if fi.Mode().IsDir() {
			items, err := GetFileItems(fPath)
			if err != nil {
				return fileItems, err
			}

			fileItems[name] = FileItem{
				Directory: true,
				Items:     items,
			}
		} else {
			chunks, err := calculateChunks(f, fi)
			if err != nil {
				return fileItems, err
			}
			fileItems[name] = FileItem{
				Size:   fi.Size(),
				Hash:   chunks.CalculateHash(),
				Chunks: chunks,
			}
		}
	}
	return fileItems, nil
}

func calculateChunks(f *os.File, fi os.FileInfo) (ChunkSlice, error) {
	nbChunks := int(math.Ceil(float64(fi.Size()) / float64(minChunkSize)))
	var chunkSize int64 = minChunkSize
	if nbChunks > maxChunksPerFile {
		nbChunks = maxChunksPerFile
		chunkSize = int64(math.Ceil(float64(fi.Size()) / float64(nbChunks)))
	}

	chunks := make(ChunkSlice, nbChunks)
	buf := make([]byte, chunkSize)

	for i := 0; i < nbChunks; i++ {
		n, err := f.Read(buf[:])
		if err != nil {
			return nil, err
		}

		hashData := md5.Sum(buf[:n])
		chunks[i] = Chunk{
			Hash: base64.URLEncoding.EncodeToString(hashData[:]),
			Size: int64(n),
		}
	}

	return chunks, nil
}
