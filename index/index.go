package index

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
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
	entries := make([]Entry, 0, len(i.Entries))
	for _, v := range i.Entries {
		entries = append(entries, v)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].name < entries[j].name
	})

	var buf bytes.Buffer

	buf.WriteString("DIRC")
	binary.Write(&buf, binary.BigEndian, i.Version)
	binary.Write(&buf, binary.BigEndian, uint32(len(i.Entries)))

	for _, e := range entries {
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
		buf.WriteByte(0)

		pad := (8 - ((buf.Len() + 20) % 8)) % 8
		for range pad {
			buf.WriteByte(0)
		}
	}

	checksum := sha1.Sum(buf.Bytes())
	buf.Write(checksum[:])

	return os.WriteFile(i.pathname, buf.Bytes(), 0o666)
}

func (i *Index) Load() error {
	data, err := os.ReadFile(i.pathname)
	if err != nil {
		return err
	}

	if len(data) < 12 {
		return fmt.Errorf("index file too small")
	}

	// parse header
	sig := string(data[0:4])

	if sig != "DIRC" {
		return fmt.Errorf("invalid index signature: %s", sig)
	}

	i.Version = binary.BigEndian.Uint32(data[4:8])
	entryLen := binary.BigEndian.Uint32(data[8:12])

	checkSum := sha1.Sum(data[:len(data)-20])
	if !bytes.Equal(checkSum[:], data[len(data)-20:]) {
		return fmt.Errorf("checkSum does not match stored value")
	}

	// clear existing Entries
	i.Entries = make(map[string]Entry, entryLen)

	offset := 12

	for range entryLen {
		if offset+62 > len(data) {
			return fmt.Errorf("truncated index entry")
		}

		e := Entry{}

		e.CTimeSec = binary.BigEndian.Uint32(data[offset : offset+4])
		e.CTimeSec = binary.BigEndian.Uint32(data[offset+4 : offset+8])
		e.MTimeSec = binary.BigEndian.Uint32(data[offset+8 : offset+12])
		e.MTimeNSec = binary.BigEndian.Uint32(data[offset+12 : offset+16])
		e.Dev = binary.BigEndian.Uint32(data[offset+16 : offset+20])
		e.Inod = binary.BigEndian.Uint32(data[offset+20 : offset+24])
		e.Mode = binary.BigEndian.Uint32(data[offset+24 : offset+28])
		e.UID = binary.BigEndian.Uint32(data[offset+28 : offset+32])
		e.GID = binary.BigEndian.Uint32(data[offset+32 : offset+36])
		e.Size = binary.BigEndian.Uint32(data[offset+36 : offset+40])

		copy([]byte(e.objHash), data[offset+40:offset+60])
		e.flags = binary.BigEndian.Uint16(data[offset+60 : offset+62])

		offset += 62

		// Read null-terminated name
		nameEnd := bytes.IndexByte(data[offset:], 0)
		if nameEnd == -1 {
			return fmt.Errorf("malformed file name in index")
		}

		e.name = string(data[offset : offset+nameEnd])
		offset += nameEnd + 1

		// Skip padding to 8-byte boundary
		pad := (8 - ((offset + 20) % 8)) % 8
		offset += pad

		i.Entries[e.name] = e
	}

	return nil
}
