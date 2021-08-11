package ipsrv

import (
    "os"
    "io"
    "fmt"
    "net"
    "math/big"
    "bytes"
    "strings"
    "encoding/binary"
    "golang.org/x/exp/mmap"
)

type IPSrvDB struct {
    Filename string
    Mode string
    M *mmap.ReaderAt
    F *os.File
    IndexSize int
    DataSize int
    IndexEnd int
    HeaderSize int
    Header string
    Date string
    Description string
    Buf []byte
    Len int
}

func (db *IPSrvDB) Open() {
    var err error
    mode := strings.ToLower(db.Mode)
    db.Mode = mode
    if mode == "mmap" {
        db.M, err = mmap.Open(db.Filename)
        db.Len = db.GetLen()
    } else if mode == "file" {
        db.F, err = os.Open(db.Filename)
        db.Len = db.GetLen()
    } else if mode == "memory" {
        db.F, err = os.Open(db.Filename)
        db.Len = db.GetLen()
        buf := make([]byte, db.Len)
        db.F.Read(buf)
        db.Buf = buf
    }

    if err != nil {
        fmt.Println(err.Error())
    } else {
        index_size := make([]byte, 8)
        data_size := make([]byte, 8)
        header_size := make([]byte, 2)
        db.ReadAt(index_size, 0)
        db.ReadAt(data_size, 8)
        db.ReadAt(header_size, 16)

        db.IndexSize = Bytes8ToInt(index_size)
        db.DataSize = Bytes8ToInt(data_size)
        db.HeaderSize = int(Bytes4ToInt(header_size))
        db.IndexEnd = 18 + db.IndexSize * 24

        data_end := db.IndexEnd + db.DataSize
        header_end := data_end + db.HeaderSize

        _header := make([]byte, header_end-data_end)
        db.ReadAt(_header, int64(data_end))
        db.Header = string(_header[:])

        _date := make([]byte, 8)
        db.ReadAt(_date, int64(header_end))
        db.Date = string(_date[:])

        _description := make([]byte, db.GetLen()-header_end-8)
        db.ReadAt(_description, int64(header_end+8))
        db.Description = string(_description[:])
    }
}

func (db *IPSrvDB) ReadAt(b []byte, off int64) (n int, err error) {
    if db.Mode == "mmap" {
        return db.M.ReadAt(b, off)
    } else if db.Mode == "file" {
        return db.F.ReadAt(b, off)
    } else if db.Mode == "memory" {
        n := copy(b, db.Buf[off:])
        if n < len(b) {
            return n, io.EOF
        }
        return n, nil
    }
    return 0, nil
}

func (db *IPSrvDB) GetLen() int {
    if db.Mode == "mmap" {
        return db.M.Len()
    } else if db.Mode == "file" || db.Mode == "memory" {
        fi, _ := db.F.Stat()
        return int(fi.Size())
    }
    return 0
}

func (db *IPSrvDB) Index(start, end int) []byte {
    b := make([]byte, end-start)
    db.ReadAt(b, int64(start+18))
    return b
}

func (db *IPSrvDB) Data(start, end int) []byte {
    b := make([]byte, end-start)
    db.ReadAt(b, int64(start+db.IndexEnd))
    return b
}

func (db *IPSrvDB) GetHeader() string {
    return db.Header
}

func (db *IPSrvDB) GetDate() string {
    return db.Date
}

func (db *IPSrvDB) GetDescription() string {
    return db.Description
}

func (db *IPSrvDB) Find(ipstr string) string {
    var ip []byte
    if strings.Index(ipstr, ":") != -1 {
        ip = net.ParseIP(ipstr).To16()
    } else {
        ip = net.ParseIP(ipstr).To4()
    }

    ipint := big.NewInt(0)
    ipint.SetBytes(ip)

    start := 0
    mid := 0
    end := db.IndexSize - 1
    for {
        if start > end {
            break
        }
        mid = int((start + end) / 2)

        high := Bytes8ToInt(db.Index(mid*24, mid*24+8))
        low := Bytes8ToInt(db.Index(mid*24+8, mid*24+16))

        unpacked := big.NewInt(int64(high))
        unpacked = unpacked.Lsh(unpacked, 64)
        unpacked = unpacked.Or(unpacked, big.NewInt(int64(low)))
        if unpacked.Cmp(ipint) > 0 {
            end = mid
        } else if unpacked.Cmp(ipint) < 0 {
            start = mid
            if start == end - 1 {
                offset0 := Bytes4ToInt(db.Index(mid*24+16, mid*24+20))
                offset1 :=  Bytes4ToInt(db.Index(mid*24+20, mid*24+24))
                info := db.Data(int(offset0), int(offset0+offset1))
                ret := string(info[:])
                return ret
            }
        } else if unpacked.Cmp(ipint) == 0 {
            offset0 := Bytes4ToInt(db.Index(mid*24+16, mid*24+20))
            offset1 :=  Bytes4ToInt(db.Index(mid*24+20, mid*24+24))
            info := db.Data(int(offset0), int(offset0+offset1))
            ret := string(info[:])
            return ret
        }
    }
    return ""
}

func (db *IPSrvDB) Findx(ipstr string) map[string]string {
    ret := make(map[string]string)
    info := db.Find(ipstr)
    header := db.GetHeader()
    info_l := strings.Split(info, ",")
    header_l := strings.Split(header, ",")
    if len(info_l) == len(header_l) {
        for i := 0; i < len(info_l); i++ {
            ret[header_l[i]] = info_l[i]
        }
    }
    return ret
}

func BytesToInt(b []byte) int {
    buf := bytes.NewBuffer(b) // b is []byte
    data, _ := binary.ReadUvarint(buf)
    return int(data)
}

func Bytes8ToInt(b []byte) int {
    var pi uint64
    buf := bytes.NewReader(b)
    binary.Read(buf, binary.LittleEndian, &pi)
    return int(pi)
}

func Bytes4ToInt(b []byte) int64 {
    xx := make([]byte, 4)
    if len(b) == 2 {
        xx = []byte{b[0], b[1], 0, 0}
    } else {
        xx = b
    }

    m := len(xx)
    nb := make([]byte, 4)
    for i := 0; i < 4; i++ {
        nb[i] = xx[m-i-1]
    }
    bytesBuffer := bytes.NewBuffer(nb)

    var x int32
    binary.Read(bytesBuffer, binary.BigEndian, &x)

    return int64(x)
}
