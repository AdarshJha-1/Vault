package wal

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
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

func (w *wal) Close() error {
	return w.currSegment.Close()
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

			// do all the SET command replay here
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

	entrySize := 4 + length
	segStat, _ := w.currSegment.Stat()
	if segStat.Size()+int64(entrySize) > w.maxFileSize {
		w.rotate()
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

// this create new seg file with all checking and set it in WAL struct
func (w *wal) rotate() error {

	// first close the current open segment file
	w.Close()

	files, err := filepath.Glob(filepath.Join(w.directory, segmentPrefix+"*"))
	if err != nil {
		return err
	}
	// increase the segment no
	w.currSegmentNo += 1

	// here i remove the starting seg no. file like if i have like this seg-0, 1, 2, 3...9 i will remove 0th one i can get it but  len(files) - maxSeg

	if len(files) == w.maxSegments {
		fileToDelete := segmentPrefix + strconv.Itoa(w.currSegmentNo-w.maxSegments)
		err := os.Remove(filepath.Join(w.directory, fileToDelete))
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
