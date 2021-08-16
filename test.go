package ipsrv

import (
    "fmt"
	"github.com/ipsrv/ipsrvdb-go"
)

func test () {
    var db ipsrv.IPSrvDB
    db.Filename = "/path/to/ipsrv.dat"
    db.Mode = "mmap"
    db.Open()
    fmt.Println(db.Find("8.8.8.255"))
    fmt.Println(db.Findx("8.8.8.255"))
    fmt.Println(db.GetHeader(), db.GetDate(), db.GetDescription())
    db.Close()
}
