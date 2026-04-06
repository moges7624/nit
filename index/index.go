package index

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"os"
	"syscall"
	"time"
)

type Index struct {
	pathname string
	Entries  map[string]Entry
	Version  uint32
}

type Stat struct{}

type Entry struct {
	objHash string
	flags   uint16
	name    string

	CTimeSec  uint32
	CTimeNSec uint32
	MTimeSec  uint32
	MTimeNSec uint32
	Dev       uint32
	Inod      uint32
	Mode      uint32
	UID       uint32
	GID       uint32
	Size      uint32
}

func NewIndex(pathname string) *Index {
	return &Index{
		pathname: pathname,
		Entries:  make(map[string]Entry),
		Version:  2,
	}
}

func (i *Index) Add(path, objHash string, stat os.FileInfo) {
	entry := i.CreateEntry(path, objHash, stat)
	i.Entries[path] = *entry
}

func (i *Index) CreateEntry(pathname, objHash string, stat os.FileInfo) *Entry {
	e := &Entry{
		name:    pathname,
		objHash: objHash,
		Mode:    0o100644,
		flags:   uint16(len(pathname) & 0x0fff),
		Size:    uint32(stat.Size()),
	}

	sys, ok := stat.Sys().(*syscall.Stat_t)
	if !ok {
		now := time.Now()
		e.CTimeSec = uint32(now.Unix())
		e.CTimeNSec = uint32(now.Nanosecond())
		e.MTimeSec = uint32(stat.ModTime().Unix())
		e.MTimeNSec = uint32(stat.ModTime().Nanosecond())
		e.Dev = 0
		e.Inod = 0
		e.UID = 0
		e.GID = 0
	}

	e.Dev = uint32(sys.Dev)
	e.Inod = uint32(sys.Ino)

	e.UID = sys.Uid
	e.GID = sys.Gid

	e.MTimeSec = uint32(sys.Mtimespec.Sec)
	e.MTimeNSec = uint32(sys.Mtimespec.Nsec)

	e.CTimeSec = uint32(sys.Ctimespec.Sec)
	e.CTimeNSec = uint32(sys.Ctimespec.Nsec)

	return e
}

func (i *Index) Write() error {
	var buf bytes.Buffer

	buf.WriteString("DIRC")
	binary.Write(&buf, binary.BigEndian, i.Version)
	binary.Write(&buf, binary.BigEndian, uint32(len(i.Entries)))

	// sort Entries

	for _, e := range i.Entries {
		binary.Write(&buf, binary.BigEndian, e.CTimeSec)
		binary.Write(&buf, binary.BigEndian, e.CTimeNSec)
		binary.Write(&buf, binary.BigEndian, e.MTimeSec)
		binary.Write(&buf, binary.BigEndian, e.MTimeNSec)
		binary.Write(&buf, binary.BigEndian, e.Dev)
		binary.Write(&buf, binary.BigEndian, e.Inod)
		binary.Write(&buf, binary.BigEndian, e.Mode)
		binary.Write(&buf, binary.BigEndian, e.UID)
		binary.Write(&buf, binary.BigEndian, e.GID)
		binary.Write(&buf, binary.BigEndian, e.Size)
		hashStr, _ := hex.DecodeString(e.objHash)
		buf.Write(hashStr)
		binary.Write(&buf, binary.BigEndian, e.flags)

		buf.WriteString(e.name)

		pad := (8 - (buf.Len() % 8)) % 8
		for range pad {
			buf.WriteByte(0)
		}
	}

	checksum := sha1.Sum(buf.Bytes())
	buf.Write(checksum[:])

	return os.WriteFile(i.pathname, buf.Bytes(), 0o666)
}
