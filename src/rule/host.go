package rule

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"rander"
	"strings"
	"util"

	"github.com/mgutz/logxi/v1"
	"gopkg.in/yaml.v2"
)

type NsRecord struct {
	Ns     string
	Name   string
	Weight int
}

type Host struct {
	Target      []Target          `json:"target"`
	Record      string            `json:"record"`
	Ttl         uint32            `json:"ttl"`
	IpKey       string            `json:"ipKey"`
	LoadBalance string            `json:"loadbalance"`
	Regex       string            `json:"regex"`
	Apiurl      string            `json:"apiurl"`
	Policy      string            `json:"policy"`
	Max         int               `json:"max"`
	Label       map[string]string `json:"label"`
	Nsrecords   []NsRecord        `json:"nsrecords"`
}

type sortTagets []Target

func (s sortTagets) Less(i, j int) bool {
	return s[i].Priority < s[j].Priority
}

func (s sortTagets) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortTagets) Len() int {
	return len(s)
}

type SortTargetsByWeight []Target

func (s SortTargetsByWeight) Less(i, j int) bool {
	return s[i].Weight > s[j].Weight
}

func (s SortTargetsByWeight) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SortTargetsByWeight) Len() int {
	return len(s)
}

var Hosts map[string]Host
var hostCache = util.NewCache(CACHE_TTL)
var rerHost *rander.Rander = rander.New()

func LoadHosts(confDir string) (err error) {
	defer func() {
		if x := recover(); x != nil {
			fmt.Println("panic:", x)
		}
	}()
	dir, er := ioutil.ReadDir(confDir)
	if er != nil {
		err = er
		return
	}
	allHosts := map[string]Host{}
	for _, file := range dir {
		if !file.IsDir() {
			if strings.HasSuffix(file.Name(), CONF_SUFFIX) {
				conf, er := ioutil.ReadFile(confDir + "/" + file.Name())
				if er != nil {
					err = er
					return
				}
				hosts := map[string]Host{}
				yaml.Unmarshal(conf, &hosts)
				for domain, host := range hosts {
					if !strings.HasSuffix(domain, ".") { //dns 协议规定必须以.结尾
						domain = domain + "."
					}
					allHosts[domain] = host
				}
				log.Info("host config loaded", "file", file.Name())
				log.Debug("host config detail", "result", fmt.Sprintf("%v", allHosts))
			}
		}
	}
	Hosts = allHosts
	log.Info("config reloaded")
	return
}

func FindHost(name string, fromRedis bool) (host Host, exist bool) {
	if fromRedis {
		return FindHostFromRedis(name)
	} else {
		return FindHostFromEtc(name)
	}
}

func FindHostFromRedis(name string) (host Host, exist bool) {
	hostInCache, ok := hostCache.Get(name)
	if ok {
		return hostInCache.(Host), true
	}
	//cache miss
	hostJson, err := client.Get("/host/" + name).Result()
	host = Host{}
	if err != nil || hostJson == "" {
		return host, false
	} else {
		log.Debug("host found in redis", "host", hostJson)
		json.Unmarshal([]byte(hostJson), &host)
		hostCache.Put(name, host)
		return host, true
	}
}

func FindHostFromEtc(name string) (host Host, exist bool) {
	host, exist = Hosts[name]
	return
}

func (host *Host) GetEntries(randomize bool, fromRedis bool) (entries []Entry) {
	return GetEntriesByTargetsFromOrigin(host.Target, randomize, fromRedis)
}
