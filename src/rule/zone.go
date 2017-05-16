package rule

import (
	"fmt"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"strings"
	"github.com/mgutz/logxi/v1"
	"encoding/json"
)

type NsInZone struct {
	Name string
	Ip string
	Weight int
}

type Zone struct {
	Origin string
	Soa string
	Ns []NsInZone
}

var Zones map[string]*Zone;

func LoadZones(confDir string)(err error){
	defer func(){
		if x := recover(); x != nil{
			fmt.Println("panic:", x)
		}
	}()
	dir, er := ioutil.ReadDir(confDir)
	if er != nil{
		err = er
		return
	}
	allZones := map[string]*Zone{}
	for _, file := range dir {
		if !file.IsDir() {
			if strings.HasSuffix(file.Name(), CONF_SUFFIX) {
				conf, er := ioutil.ReadFile(confDir + "/" + file.Name())
				if er != nil {
					err = er
					return
				}
				zones := map[string]*Zone{}
				yaml.Unmarshal(conf, &zones)
				for domain, zone := range zones {
					if !strings.HasSuffix(domain, ".") {//dns 协议规定必须以.结尾
						domain = domain + "."
					}
					allZones[domain] = zone
				}
				log.Info("host config loaded", "file", file.Name())
				log.Debug("host config detail", "result", fmt.Sprintf("%v", allZones))
			}
		}
	}
	Zones = allZones
	log.Info("config reloaded")
	return
}

/**
name: 域名
equal: 需要精确一致
 */
func FindZone(name string, equal bool)(zone *Zone){
	log.Debug("find zone", "name", name)
	if !ruleConf.FromRedis {
		zone = findZoneFromMem(name, equal)
	}else {
		zone = findZoneFromRedis(name, equal)
	}
	return
}

func findZoneFromMem(name string, equal bool)(zone *Zone) {
	if equal {
		zone, _ = Zones[name]
	}else {
		atoms := strings.Split(name, ".")
		l := len(atoms)
		for i := l - 3; i >= 0 && i > l - 6; i-- {
			//模糊搜索，最大探索4层
			domain := ""
			for j := l - 2; j >= i; j-- {
				domain = atoms[j] + "." + domain
			}
			var ok bool;
			zone, ok = Zones[name]
			if ok {
				break
			}
		}
	}
	return
}

func findZoneFromRedis(name string, equal bool)(zone *Zone) {
	if equal {
		zoneJson, err := client.Get("/zones/" + name).Result()
		if err == nil || zoneJson != "" {
			log.Debug("zone found", "origin", name, "value", zoneJson)
			zone = &Zone{}
			json.Unmarshal([]byte(zoneJson), zone)
		}
	}else {
		atoms := strings.Split(name, ".")
		l := len(atoms)
		for i := l - 3; i >= 0 && i > l - 6; i-- {
			//模糊搜索，最大探索4层
			domain := ""
			for j := l - 2; j >= i; j-- {
				domain = atoms[j] + "." + domain
			}
			zoneJson, err := client.Get("/zones/" + domain).Result()
			if err == nil || zoneJson != "" {
				log.Debug("zone found", "origin", domain, "value", zoneJson)
				zone = &Zone{}
				json.Unmarshal([]byte(zoneJson), zone)
				break
			}
		}
	}
	return
}