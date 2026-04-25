package upload_file

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// TestExpectedVideoChunkSize 执行业务处理。
func TestExpectedVideoChunkSize(t *testing.T) {
	meta := videoChunkUploadMeta{
		FileSize:    5*1024 + 321,
		ChunkSize:   2 * 1024,
		TotalChunks: 3,
	}

	if size := expectedVideoChunkSize(meta, 0); size != 2*1024 {
		t.Fatalf("expected first chunk size 2048, got %d", size)
	}
	if size := expectedVideoChunkSize(meta, 1); size != 2*1024 {
		t.Fatalf("expected second chunk size 2048, got %d", size)
	}
	if size := expectedVideoChunkSize(meta, 2); size != 1345 {
		t.Fatalf("expected last chunk size 1345, got %d", size)
	}
}

// TestValidateVideoChunkHash 执行业务处理。
func TestValidateVideoChunkHash(t *testing.T) {
	validHash := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if err := validateVideoChunkHash(validHash); err != nil {
		t.Fatalf("expected valid hash, got error: %v", err)
	}

	if err := validateVideoChunkHash("bad-hash"); err == nil {
		t.Fatal("expected invalid hash to fail")
	}
}

// TestListValidUploadedChunkIndexes 执行业务处理。
func TestListValidUploadedChunkIndexes(t *testing.T) {
	tempDir := t.TempDir()
	meta := videoChunkUploadMeta{
		FileSize:    10,
		ChunkSize:   4,
		TotalChunks: 3,
	}

	writeChunk := func(name string, size int) {
		t.Helper()
		data := make([]byte, size)
		if err := os.WriteFile(filepath.Join(tempDir, name), data, 0644); err != nil {
			t.Fatalf("failed to write chunk %s: %v", name, err)
		}
	}

	writeChunk(videoChunkFileName(0), 4)
	writeChunk(videoChunkFileName(1), 4)
	writeChunk(videoChunkFileName(2), 1)
	writeChunk("chunk-999999.part", 4)
	writeChunk("random.txt", 4)

	indexes, err := listValidUploadedChunkIndexes(meta, tempDir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := []int{0, 1}
	if !reflect.DeepEqual(indexes, expected) {
		t.Fatalf("expected indexes %v, got %v", expected, indexes)
	}
}

// TestCollectVerifiedVideoChunks 执行业务处理。
func TestCollectVerifiedVideoChunks(t *testing.T) {
	tempDir := t.TempDir()
	meta := videoChunkUploadMeta{
		FileSize:    10,
		ChunkSize:   4,
		TotalChunks: 3,
	}

	chunkData := map[int][]byte{
		0: []byte("abcd"),
		1: []byte("efgh"),
		2: []byte("ij"),
	}

	for chunkIndex, data := range chunkData {
		if err := os.WriteFile(filepath.Join(tempDir, videoChunkFileName(chunkIndex)), data, 0644); err != nil {
			t.Fatalf("failed to write chunk %d: %v", chunkIndex, err)
		}
	}

	validHash := sha256.Sum256(chunkData[0])
	verifiedChunks, verifiedHashes, invalidChunks, err := collectVerifiedVideoChunks(meta, tempDir, map[int]string{
		0: hex.EncodeToString(validHash[:]),
		1: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !reflect.DeepEqual(verifiedChunks, []int{0, 2}) {
		t.Fatalf("expected verified chunks [0 2], got %v", verifiedChunks)
	}

	if !reflect.DeepEqual(invalidChunks, []int{1}) {
		t.Fatalf("expected invalid chunks [1], got %v", invalidChunks)
	}

	if verifiedHashes[0] != hex.EncodeToString(validHash[:]) {
		t.Fatalf("expected chunk 0 hash to be preserved, got %s", verifiedHashes[0])
	}

	lastHash := sha256.Sum256(chunkData[2])
	if verifiedHashes[2] != hex.EncodeToString(lastHash[:]) {
		t.Fatalf("expected chunk 2 hash to be calculated, got %s", verifiedHashes[2])
	}
}
