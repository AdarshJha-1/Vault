package wal

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func createNewSegmentFile(directory string, segmentID int) (*os.File, error) {
	filePath := filepath.Join(directory, fmt.Sprintf("segment-%d", segmentID))
	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func findLastSegmentID(files []string) (int, error) {
	var lastSegmentID int
	for _, file := range files {
		_, fileName := filepath.Split(file)
		segmentID, err := strconv.Atoi(strings.TrimPrefix(fileName, "segment-"))
		if err != nil {
			return 0, err
		}
		if segmentID > lastSegmentID {
			lastSegmentID = segmentID
		}
	}
	return lastSegmentID, nil
}

func createCRC(lastSequenceNo uint64, command string) uint32 {

	var lsnBuf [8]byte
	binary.LittleEndian.PutUint64(
		lsnBuf[:],
		lastSequenceNo,
	)

	crc := crc32.ChecksumIEEE(
		append(lsnBuf[:], []byte(command)...),
	)
	return crc
}

func verifyCRC(lastSequenceNo uint64, command string, actualCRC uint32) bool {
	var lsnBuf [8]byte
	binary.LittleEndian.PutUint64(
		lsnBuf[:],
		lastSequenceNo,
	)

	crc := crc32.ChecksumIEEE(
		append(lsnBuf[:], []byte(command)...),
	)
	return crc == actualCRC
}
