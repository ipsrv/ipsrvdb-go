# ipsrvdb-go

# Feature
1. Support IPv4 & IPv6.
2. Support output db date, description and header.
3. Support output raw IP info and IP info in a map.
4. Support load the database into memory or using MMAP.

# Installing
```
go get github.com/ipsrv/ipsrvdb-go
```

# Example
```
package main

import (
	"fmt"
	"github.com/ipsrv/ipsrvdb-go"
)

func main() {
	var db ipsrv.IPSrvDB
    db.Filename = "/path/to/ipsrv.dat"
    db.Mode = "mmap"
    db.Open()
    fmt.Println(db.Find("8.8.8.255"))
    fmt.Println(db.Findx("8.8.8.255"))
    fmt.Println(db.GetHeader(), db.GetDate(), db.GetDescription())
    db.Close()
}
```

# Output
```
NA,北美洲,US,美国,,,,,,
map[country_iso_code:US isp_zh: country_zh:美国 province_iso_code: province_zh: city_code: city_zh: org: continent_code:NA continent_zh:北美洲]
continent_code,continent_zh,country_iso_code,country_zh,province_iso_code,province_zh,city_code,city_zh,isp_zh,org 20210811 IPSrv, Inc. Dat database.
```
