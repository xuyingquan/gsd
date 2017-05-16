package rule

import (
    "github.com/mgutz/logxi/v1"
    "fmt"
    "os"
    "bufio"
    "io"
    "strings"
    "strconv"
)

var UNKNOWN_REGION = Region{"未知","未知", 0, 0}

type IPDB struct {
    regions []Region
}

type Region struct{//ip转换为int64（考虑兼容ipv4和ipv6）
    Region  string;//区域名称
    Isp string;//isp名称
    start int64; //起始ip
    end   int64; //终止ip
}

func (db *IPDB)FindRegionByIp(ip int64)(region Region){
    if len(db.regions) < 1 {
        return UNKNOWN_REGION
    }
    startFlag := 0
    stopFlag := len(db.regions) - 1
    i := 0

    for startFlag <= stopFlag && i < 1000 {
        i += 1
        middleFlag := (startFlag + stopFlag) / 2
        if (ip >= db.regions[middleFlag].start && ip <= db.regions[middleFlag].end) {
            return db.regions[middleFlag]
        }

        if (ip < db.regions[middleFlag].start) {
            stopFlag = middleFlag-1
            continue
        }

        if (ip > db.regions[middleFlag].end) {
            startFlag = middleFlag+1
        }
    }
    return UNKNOWN_REGION
}

func (db *IPDB)LoadIPDB(dbFile string){
    log.Info("load ipdb", "file", dbFile)
    tmpRegions := []Region{}
    file, err := os.Open(dbFile)
    if err != nil {
        log.Warn("open ipdb failed", "file", file, "err", fmt.Sprintf("%v", err))
        return
    }
    reader := bufio.NewReader(file)
    for {
        line, _, err := reader.ReadLine()
        if err != nil {
            if err != io.EOF {
                log.Error("read ipdb failed", "file", file, "err", fmt.Sprintf("%v", err))
                return
            }else{
                break
            }
        }
        s := fmt.Sprintf("%s", line)
        columns := strings.Split(s, ",")
        if len(columns) < 4 {
            log.Error("wrong ipdb syntx", "line", s)
            return
        }
        startIp, err := strconv.ParseInt(columns[0], 10, 0)
        endIp, err := strconv.ParseInt(columns[1], 10, 0)
        region := strings.TrimSpace(columns[2])
        isp := strings.TrimSpace(columns[3])
        if err  != nil {
            log.Error("wrog ipdb syntx", "line", s)
            return
        }
        tmpRegions = append(tmpRegions, Region{start: startIp, end: endIp, Region: region, Isp: isp})
    }
    db.regions = tmpRegions[:]
    log.Info("ipdb loaded", "length", len(db.regions))
    return
}