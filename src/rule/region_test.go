package rule

import (
    "testing"
    "time"
    "fmt"
    "net"
    "encoding/binary"
)

func TestFindRegionById(t *testing.T){
    ipdb = &IPDB{}
    ipdb.LoadIPDB("/Users/chenggong/Documents/shataspace/nsd/etc/ipdb")
    n := 100000;
    start := time.Now()
    for i := 0; i < n; i++ {
        r := ipdb.FindRegionByIp(int64(i + 3658820673))
//        r := FindRegionByIp(int64(i))
        if i%1000 == 0 {
            fmt.Println(r)
        }
    }
    end := time.Now()
    fmt.Printf("search benchmark: %dk op/s\n", int64(n)/((end.UnixNano() - start.UnixNano())/1000/1000))
}

func TestFindRegionSh(t *testing.T){
    ip := net.ParseIP("101.95.31.62")
    ipv4 := ip.To4()
    r := ipdb.FindRegionByIp(int64(binary.BigEndian.Uint32(ipv4[:])))
    fmt.Println("region:", r)
}
