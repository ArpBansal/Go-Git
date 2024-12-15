package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strconv"
)

type GitCommit struct {
	BaseGitObject
}

type GitTree struct {
	BaseGitObject
}

type GitTag struct {
	BaseGitObject
}

type GitBlob struct {
	BaseGitObject
}

func NewGitObject(data bytes.Buffer) (*GitObject, error) {
	return nil, nil
}

func (gitobj *BaseGitObject) Serialize(repo *GitRepository) error {
	return nil
}

func (gitobj *BaseGitObject) Deserialize(data bytes.Buffer) error {
	return nil
}

func (gitobj *BaseGitObject) Format() string {
	return ""
}

func (gitobj *BaseGitObject) Init() {

}

func (repo *GitRepository) objectRead(sha string) (GitObject, error) {
	path := repo.RepoFile(false, "objects", sha[0:2], sha[2:])
	if path == "" {
		return nil, fmt.Errorf("object not found")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open object file: %v", err)
	}
	defer file.Close()

	zr, err := zlib.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("could not create zlib reader: %v", err)
	}
	defer zr.Close()

	var raw bytes.Buffer
	if _, err := io.Copy(&raw, zr); err != nil {
		return nil, fmt.Errorf("could not decompress object: %v", err)
	}

	// Read object type
	x := bytes.IndexByte(raw.Bytes(), ' ')
	if x == -1 {
		return nil, fmt.Errorf("malformed object: missing space")
	}
	fmtType := raw.Bytes()[:x]

	// Read and validate object size
	y := bytes.IndexByte(raw.Bytes()[x:], '\x00')
	if y == -1 {
		return nil, fmt.Errorf("malformed object: missing null byte")
	}
	y += x
	size, err := strconv.Atoi(string(raw.Bytes()[x+1 : y]))
	if err != nil {
		return nil, fmt.Errorf("malformed object: invalid size")
	}
	if size != len(raw.Bytes())-y-1 {
		return nil, fmt.Errorf("malformed object %s: bad length", sha)
	}

	// Pick constructor
	var obj GitObject
	switch string(fmtType) {
	case "commit":
		obj = &GitCommit{}
	case "tree":
		obj = &GitTree{}
	case "tag":
		obj = &GitTag{}
	case "blob":
		obj = &GitBlob{}
	default:
		return nil, fmt.Errorf("unknown type %s for object %s", fmtType, sha)
	}

	// Call constructor and return object
	if err := obj.Deserialize(*bytes.NewBuffer(raw.Bytes()[y+1:])); err != nil {
		return nil, fmt.Errorf("could not deserialize object: %v", err)
	}
	return obj, nil
}

// func (repo *GitRepository) ObjectRead(sha string) (*GitObject, error) {
// 	path := repo.RepoFile("objects", sha[0:2], sha[2:])

// 	if !fileExists(path) {
// 		return nil, fmt.Errorf("object not found")
// 	}

// 	raw, err := decompressFile(path)
// 	if err != nil {
// 		return nil, err
// 	}

// 	x := bytes.IndexByte(raw, ' ')
// 	fmt := raw[:x]

// 	y := bytes.IndexByte(raw[x:], '\x00') + x
// 	size, err := strconv.Atoi(string(raw[x+1 : y]))
// 	if err != nil || size != len(raw)-y-1 {
// 		return nil, fmt.Errorf("malformed object %s: bad length", sha)
// 	}

// 	var obj GitObject
// 	switch string(fmt) {
// 	case "commit":
// 		obj = &GitCommit{}
// 	case "tree":
// 		obj = &GitTree{}
// 	case "tag":
// 		obj = &GitTag{}
// 	case "blob":
// 		obj = &GitBlob{}
// 	default:
// 		return nil, fmt.Errorf("unknown type %s for object %s", fmt, sha)
// 	}

// 	err = obj.Deserialize(raw[y+1:])
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &obj, nil
// }

func objectWrite(obj GitObject, repo *GitRepository) (string, error) {
	// Serialize object data
	data, err := obj.Serialize()
	if err != nil {
		return "", fmt.Errorf("failed to serialize object: %v", err)
	}

	// Add header
	header := fmt.Sprintf("%s %d\x00", obj.Format(), len(data))
	result := append([]byte(header), data...)

	// Compute hash
	sha := fmt.Sprintf("%x", sha256.Sum256(result))

	if repo != nil {
		// Compute path
		path := repo.RepoFile(true, "objects", sha[:2], sha[2:])
		if path == "" {
			return "", fmt.Errorf("failed to compute object path")
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			// Compress and write
			var buf bytes.Buffer
			zw := zlib.NewWriter(&buf)
			if _, err := zw.Write(result); err != nil {
				return "", fmt.Errorf("failed to compress object: %v", err)
			}
			if err := zw.Close(); err != nil {
				return "", fmt.Errorf("failed to close zlib writer: %v", err)
			}

			if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
				return "", fmt.Errorf("failed to write object file: %v", err)
			}
		}
	}

	return sha, nil
}
