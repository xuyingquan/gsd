package rule
import (
    "net"
    "github.com/mgutz/logxi/v1"
    "fmt"
    "sync"
    "time"
    "strings"
    "regexp"
    "net/http"
    "io/ioutil"
    "encoding/binary"
    "gopkg.in/redis.v3"
    "gopkg.in/yaml.v2"
)

const CONF_SUFFIX = "yaml"

const GEOIP = "GEOIP"
const FIXED = "FIXED"

const TypeA = "TYPEA"
const TypeCNAME = "TypeCNAME"

const LB_CONSISTEN_HASH = "consistent_hash"
const LB_SDN_API = "sdn_api"
const LB_POLICY = "policy"
const LB_RANDOM = "random"
const LB_ALL = "all"

var reloadLock sync.RWMutex
var confPath = "etc"

type Conf struct {
    FromRedis bool `yaml:"redis"`
    RedisAddr string `yaml:"redisAddr"`
    RedisPass string `yaml:"redisPass"`
    CacheExpires time.Duration `yaml:"cacheExpires"`
}

type Record struct {
    RecordType string
    Ttl uint32
    Ip net.IP
    CName string
}

var ipdb *IPDB
var priorIpdb *IPDB
var ruleConf Conf

var client *redis.Client

func InitRedis(addr string, password string)(err error){
    client = redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: password, // no password set
        DB:       0,  // use default DB
    })
    _, err = client.Ping().Result()
    if err == nil {
        log.Info("connect to redis ok", "addr", addr)
    }
    return
}

func LoadIPDB(confDir string){
    ipdb = &IPDB{}
    priorIpdb = &IPDB{}
    ipdb.LoadIPDB(confDir + "/ipdb")
    priorIpdb.LoadIPDB(confDir + "/pipdb")
}

func LoadConf(confDir string){
    confPath = confDir
    reloadLock.Lock()
    defer reloadLock.Unlock()
    LoadIPDB(confDir)
    conf, err := ioutil.ReadFile(confDir + "/gsd.yaml")
    if err != nil {
        log.Warn("reload gsd conf failed", "err", fmt.Sprintf("%v", err))
    }
    yaml.Unmarshal(conf, &ruleConf)
    log.Info("gsd config loaded", "conf", ruleConf)
    if ruleConf.FromRedis {
        err := InitRedis(ruleConf.RedisAddr, ruleConf.RedisPass)
        if err != nil {
            log.Error("connect redis failed", "err", fmt.Sprintf("%v", err))
        }
    }else {
        LoadPools(confDir + "/pools")
        LoadHosts((confDir + "/hosts"))
        LoadPolicies(confDir + "/policies")
        LoadZones(confDir + "/zones")
    }

    if ruleConf.CacheExpires != 0 {
        CacheExpiredInterval = ruleConf.CacheExpires
    }
}

func Reload(){
    LoadConf(confPath)
}

func FindRecords(name string, remoteIp net.IP, ignoreMax bool)(records *[]Record, nsRecords []NsRecord, labels map[string]string){
    reloadLock.RLock()
    defer reloadLock.RUnlock()
    domain := name
    log.Debug("find records by", "name", name, "remoteIp", remoteIp.String())
    max := 0
FIND_RECORD:
    for i := 0;i<10;i++{//前缀模糊匹配，最大探索10层
        host, ok := FindHost(domain, ruleConf.FromRedis)
        if ok {
            log.Debug("host conf found", "name", name, "domain", domain, "host", fmt.Sprintf("%v", host))
            //todo: 实现偏简单，可以进一步优化
            if len(host.Nsrecords) > 0 {
                nsRecords = rerHost.Randomize(host.Nsrecords).([]NsRecord)
                log.Debug("host has ns record", "records", nsRecords)
            }
            labels = host.Label
            max = host.Max
            switch host.LoadBalance {
                case LB_SDN_API://基于SDN接口，返回接口返回值构成的记录
                log.Debug("use sdn rule")
                reg, err := regexp.Compile(host.Regex)
                if err != nil {
                    log.Error("can't compile regex", "string", host.Regex)
                    break FIND_RECORD
                }
                url := strings.Replace(host.Apiurl, "$remoteIp", remoteIp.String(), -1)
                url = reg.ReplaceAllString(name, url)
                if url[len(url)-1] == '.' {//去掉最后那个.
                    url = url[:len(url)-1]
                }
                records = getRecordsFromSDN(url, host)
                break FIND_RECORD

                case LB_POLICY://基于POLICY配置，按用户来源分配
                log.Debug("use policy rule")
                var matchRegion string
                records, matchRegion = getRecordsByPolicy(remoteIp, domain ,host)
                if labels == nil {
                    labels = map[string]string{}
                }
                labels["region"] = matchRegion
                break FIND_RECORD

                case LB_RANDOM://基于target pool，随机打散
                records = getRecordsAll(domain, host, true)
                break FIND_RECORD

                default:// return first record
                log.Debug("use all rule")
                records = getRecordsAll(domain, host, false)
                break FIND_RECORD
            }
        }else {
            //依次模糊匹配
            if domain == "." {
                break
            }
            p := strings.Index(domain[1:], ".")
            if p < 0 {
                domain = "."
            }
            domain = domain[p+1:]
        }
    }
    if !ignoreMax && max != 0 && records != nil && len(*records) >= max {
        result := (*records)[:max]
        records = &result
    }
    return
}

func getRecordsAll(domain string, host Host, randomize bool)(records *[]Record){
    log.Debug("find record by rule all",  "pools", host.Target, "random", randomize)
    entries := host.GetEntries(randomize, ruleConf.FromRedis)
    log.Debug("the entries found ", "entries", entries)
    if entries != nil {
        length := len(entries)
        if length > 0 {
            records = &[]Record{}
            for _, entry := range entries {
                record := genRecord(entry, host)
                if record != nil{
                    *records = append(*records, *record)
                }
            }
        }
    }
    return
}

func getRecordsFromSDN(url string, host Host)(records *[]Record){
    log.Debug("sdn api call url", "url", url)
    resp, err := http.Get(url)
    if err != nil {
        log.Error("sdn api call failed", "url", url, "err", fmt.Sprintf("%v", err))
        return
    }
    defer resp.Body.Close()
    if resp.StatusCode != 200 {
        log.Error("sdn api reponse err", "status", resp.Status)
        return
    }
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Error("sdn api read reponse body faild", "err", fmt.Sprintf("%v", err))
        return
    }
    ip := string(body)
    rip := net.ParseIP(ip)
    if err != nil || rip == nil {
        log.Error("sdn api parse response ip failed", "ip", ip, "err", fmt.Sprintf("%v", err))
    }
    records = &[]Record{Record{RecordType: TypeA, Ttl: host.Ttl, Ip: rip}}
    return
}

func getRecordsByPolicy(remoteIp net.IP, domain string, host Host)(records *[]Record, matchRegion string){
    log.Debug("find record by remote ip", "remoteIP", fmt.Sprintf("%v", remoteIp), "domain", domain, "host", fmt.Sprintf("%v", host))
    ip32 := binary.BigEndian.Uint32(remoteIp.To4()[:])
    region := priorIpdb.FindRegionByIp(int64(ip32))
    if region.start == 0 && region.end == 0{//no match region
        region = ipdb.FindRegionByIp(int64(ip32))
    }
    entries, matchRegion := findEntriesByRegion(host.Policy, region, domain, ruleConf.FromRedis)
    if entries != nil {
        records = &[]Record{}
        for _, entry := range entries {
            record := genRecord(entry, host)
            if record != nil {
                *records = append(*records, *record)
            }
        }
    }
    return
}

func genRecord(entry Entry, host Host)(record *Record){
    if host.Record == "A" {
        ip := entry.Ip[host.IpKey]
        rip := net.ParseIP(ip)
        if rip == nil {
            log.Error("wrong config ip", "ip", ip)
        }
        record = &Record{RecordType: TypeA, Ttl: host.Ttl, Ip: rip}
    }
    if host.Record == "CNAME" {
        cname := entry.Cname
        if cname != "" {
            if cname[len(cname) - 1] != '.' {
                cname = cname + "."
            }
            record = &Record{RecordType: TypeCNAME, Ttl: host.Ttl, CName: cname}
        }
    }
    return
}
