package wal

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	// here i get all present segment files
	files, err := filepath.Glob(filepath.Join(directory, segmentPrefix+"*"))

	// zero value is 0 so its good
	var lastSegmentID int
	// if no segment files is present create a first file and set in WAL struct
	if len(files) == 0 {
		file, err := createNewSegmentFile(directory, 0)
		if err != nil {
			return nil, err
		}
		if err := file.Close(); err != nil {
			return nil, err
		}
	} else { // if present get the last one and set in WAL struct so we can continue from it
		lastSegmentID, err = findLastSegmentID(files)
		if err != nil {
			return nil, err
		}
	}

	// now i just open the currSegmentFile
	filePath := filepath.Join(directory, fmt.Sprintf("%s%d", segmentPrefix, lastSegmentID))
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	// seek to end of the file so appending can proceed from there only
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

// i think WAL can be standalone thing but i am not that smart i am mixing all up in one code only no segregation
func (w *wal) LoadToVault(storage store.Store) error {
	files, err := filepath.Glob(filepath.Join(w.directory, segmentPrefix+"*"))
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return nil
	}
	for _, file := range files {
		// read each file content verify crc checksum if valid add in vault other wise go to next file
		// fmt.Println(file)

		openFile, err := os.OpenFile(file, os.O_RDONLY, 0644)
		if err != nil {
			return err
		}
		defer openFile.Close()
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

			fmt.Printf("%+v\n", entry)
		}
	}
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
