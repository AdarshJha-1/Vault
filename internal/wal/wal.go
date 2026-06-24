package wal

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/AdarshJha-1/Vault/internal/store"
	walpb "github.com/AdarshJha-1/Vault/proto/wal"
	"google.golang.org/protobuf/proto"
)

const (
	segmentPrefix = "segment-"
)

type WAL interface {
	LoadToVault(storage store.Store) error
	WriteEntry(command string)
	Close() error
}

type wal struct {
	directory      string
	currSegment    *os.File
	lock           sync.RWMutex
	lastSequenceNo uint64
	maxFileSize    int64
	maxSegments    int
	currSegmentNo  int
}

func OpenWAL(directory string, maxFileSize int64, maxSegments int) (WAL, error) {
	if err := os.MkdirAll(directory, 0755); err != nil {
		return nil, err
	}
	files, err := filepath.Glob(filepath.Join(directory, segmentPrefix+"*"))

	var lastSegmentID int
	if len(files) == 0 {
		file, err := createNewSegmentFile(directory, 0)
		if err != nil {
			return nil, err
		}
		if err := file.Close(); err != nil {
			return nil, err
		}
	} else {
		lastSegmentID, err = findLastSegmentID(files)
		if err != nil {
			return nil, err
		}
	}

	filePath := filepath.Join(directory, fmt.Sprintf("%s%d", segmentPrefix, lastSegmentID))
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	if _, err := file.Seek(0, io.SeekEnd); err != nil {
		return nil, err
	}

	return &wal{
		directory:      directory,
		currSegment:    file,
		lock:           sync.RWMutex{},
		lastSequenceNo: 0,
		maxFileSize:    maxFileSize,
		maxSegments:    maxSegments,
		currSegmentNo:  lastSegmentID,
	}, nil
}

func (w *wal) Close() error {
	return w.currSegment.Close()
}

func (w *wal) LoadToVault(storage store.Store) error {
	files, err := filepath.Glob(filepath.Join(w.directory, segmentPrefix+"*"))
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return nil
	}

	// TODO have to change this sort thing
	sort.Strings(files)

	var maxLSN uint64 = 0
	for _, file := range files {

		openFile, err := os.Open(file)
		if err != nil {
			return err
		}
		for {
			var length uint32
			err = binary.Read(openFile, binary.LittleEndian, &length)
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}

			buf := make([]byte, length)

			_, err = io.ReadFull(openFile, buf)
			if err != nil {
				if err == io.EOF || err == io.ErrUnexpectedEOF {
					break
				}
				return err
			}
			var entry walpb.WALEntry

			err = proto.Unmarshal(buf, &entry)
			if err != nil {
				return err
			}

			if !verifyCRC(entry.GetLsn(), entry.GetCommand(), entry.GetCrc()) {
				break
			}

			commandParts := strings.Fields(entry.GetCommand())
			if len(commandParts) >= 3 {
				if strings.ToUpper(commandParts[0]) == "SET" {
					storage.Set(commandParts[1], commandParts[2])
				} else if strings.ToUpper(commandParts[0]) == "DEL" {
					storage.Delete(commandParts[1])
				}
			}
			maxLSN = max(maxLSN, entry.GetLsn())
		}

		openFile.Close()
	}
	w.lastSequenceNo = maxLSN
	return nil
}

func (w *wal) WriteEntry(command string) {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.lastSequenceNo++

	crc := createCRC(w.lastSequenceNo, command)
	entry := &walpb.WALEntry{
		Lsn:     w.lastSequenceNo,
		Command: command,
		Crc:     crc,
	}

	entryBytes, err := proto.Marshal(entry)
	if err != nil {
		return
	}
	length := uint32(len(entryBytes))

	entrySize := 4 + length
	segStat, _ := w.currSegment.Stat()
	if segStat.Size()+int64(entrySize) > w.maxFileSize {
		err := w.rotate()
		if err != nil {
			return
		}
	}

	err = binary.Write(w.currSegment, binary.LittleEndian, length)
	if err != nil {
		return
	}

	_, err = w.currSegment.Write(entryBytes)
	if err != nil {
		return
	}

	if err := w.currSegment.Sync(); err != nil {
		return
	}

}

func (w *wal) rotate() error {

	if err := w.Close(); err != nil {
		return err
	}

	files, err := filepath.Glob(filepath.Join(w.directory, segmentPrefix+"*"))
	if err != nil {
		return err
	}

	w.currSegmentNo += 1

	// TODO here too
	sort.Strings(files)

	if len(files) == w.maxSegments {
		err := os.Remove(files[0])
		if err != nil {
			return err
		}
	}
	newSeg, err := createNewSegmentFile(w.directory, w.currSegmentNo)
	if err != nil {
		return err
	}

	err = newSeg.Close()
	if err != nil {
		return err
	}

	filePath := filepath.Join(w.directory, fmt.Sprintf("%s%d", segmentPrefix, w.currSegmentNo))
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	if _, err := file.Seek(0, io.SeekEnd); err != nil {
		return err
	}

	w.currSegment = file
	return nil
}
