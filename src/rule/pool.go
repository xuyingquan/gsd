package rule

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"rander"
	"sort"
	"strings"
	"util"
	"sync"
	"crypto/md5"
	"time"

	"github.com/mgutz/logxi/v1"
	"gopkg.in/yaml.v2"
)

var Pools map[string]Pool

var (
	MatrixCache *SafeMatrixMap = &SafeMatrixMap{M: make(map[string]*Matrix, 0)} // 缓存分解矩阵
	AccessHistoryCache *SafeAccessHistoryMap = &SafeAccessHistoryMap{M: make(map[string]*AccessHistory)}
	CacheExpiredInterval time.Duration = 24 * 3600 // 清理MatrixCache和AccessHistoryCache过期时间间隔
)

// 定期清理MatrixCache和AccessHistoryCache
func CleanCache() {
	defer func() {
		if err := recover(); err != nil {
			log.Error("CleanCache", "recover", err)
		}
	}()

	for {
		time.Sleep(CacheExpiredInterval * time.Second)
		MatrixCache.Lock()
		for key, matrix := range MatrixCache.M {
			if time.Now().Sub(matrix.AccessTime) > CacheExpiredInterval {
				delete(MatrixCache.M, key)
			}
		}
		MatrixCache.Unlock()

		AccessHistoryCache.Lock()
		for key, accessHistory := range AccessHistoryCache.M {
			if time.Now().Sub(accessHistory.AccessTime) > CacheExpiredInterval {
				delete(AccessHistoryCache.M, key)
			}
		}
		AccessHistoryCache.Unlock()
	}
}

type Pool struct {
	Entries []Entry
	Disable bool
}

type Entry struct {
	Name    string
	Ip      map[string]string
	Cname   string
	Weight  int
	Disable bool
}

type AccessHistory struct {
	AccessMatrixIndex int		// 记录访问分解矩阵索引
	AccessPoolIndex []int		// 记录每个pool取IP索引
	AccessTime time.Time 		// 访问时间
}

type SafeAccessHistoryMap struct {
	M map[string]*AccessHistory
	sync.RWMutex
}

func (this *SafeAccessHistoryMap) GetAccessHistory(key string) (historyRecord *AccessHistory) {
	this.RLock()
	defer this.RUnlock()

	value, ok := this.M[key]
	if ok {
		this.M[key].AccessTime = time.Now()
		return value
	} else {
		return nil
	}
}

func (this *SafeAccessHistoryMap) SetAccessHistory(key string, historyRecord *AccessHistory) {
	this.Lock()
	defer this.Unlock()

	historyRecord.AccessTime = time.Now()
	this.M[key] = historyRecord
}


type Matrix struct {
	Data [][]int			// 分解矩阵二维数组
	IpnumList []int			// Ip节点个数列表
	WeightList []int		// 权重列表
	AccessTime time.Time 	// 访问时间
}

type SafeMatrixMap struct {
	M map[string]*Matrix // string为md5(ipnumList+weightList)
	sync.RWMutex		 // 读写锁
}

func (this *SafeMatrixMap) SetMatrix(key string, matrix *Matrix) {
	this.Lock()
	defer this.Unlock()

	matrix.AccessTime = time.Now()
	this.M[key] = matrix
}

func (this *SafeMatrixMap) GetMatrix(key string) (matrix *Matrix) {
	this.RLock()
	defer this.RUnlock()

	value, ok := this.M[key]
	if ok {
		this.M[key].AccessTime = time.Now()
		return value
	} else {
		return nil
	}
}

func GetPoolsByTarget(availableTargets []Target, fromRedis bool) (pools []Pool, ips []int, weights []int) {
	pools = make([]Pool, 0)
	weights = make([]int, 0)
	ips = make([]int, 0)

	for _, target := range availableTargets {
		entriesFound := FindEntriesByName(target.Pool, false, fromRedis)

		if target.Weight == 0 {
			target.Weight = 1
		}
		pool := Pool{Entries: entriesFound}
		pools = append(pools, pool)
		ips = append(ips, len(entriesFound))
		weights = append(weights, target.Weight)
	}
	log.Debug("GetPoolsByTarget", "available pools", pools)
	return
}

var poolCache = util.NewCache(1)
var rerPool *rander.Rander = rander.New()

func LoadPools(confDir string) (err error) {
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
	allPools := map[string]Pool{}
	for _, file := range dir {
		if !file.IsDir() {
			if strings.HasSuffix(file.Name(), CONF_SUFFIX) {
				conf, er := ioutil.ReadFile(confDir + "/" + file.Name())
				if er != nil {
					err = er
					return
				}
				pools := map[string][]Entry{}
				yaml.Unmarshal(conf, &pools)
				log.Info("pool config loaded", "file", file.Name())
				log.Debug("pool config detail", "result", fmt.Sprintf("%v\n", pools))
				for k, v := range pools {
					allPools[k] = Pool{Entries: v}
				}
			}
		}
	}
	Pools = allPools
	return
}

func pickAvailableEntries(entries []Entry) (result []Entry) {
	result = []Entry{}
	if entries == nil {
		return
	}
	for _, entry := range entries {
		if !entry.Disable {
			result = append(result, entry)
		}
	}
	return result
}

func FindEntriesByName(name string, randomize bool, fromRedis bool) (entries []Entry) {
	if fromRedis {
		entries = FindEntriesByNameFromRedis(name)
	} else {
		entries = findEntriesByNameFromEtc(name)
	}
	log.Debug("pool found", "name", name, "entries", entries)
	entries = pickAvailableEntries(entries)
	if randomize {
		entries = rerPool.Randomize(entries).([]Entry)
	}
	return entries
}

func FindEntriesByNameFromRedis(name string) (entries []Entry) {
	poolInCache, ok := poolCache.Get(name)
	if ok {
		pool := poolInCache.(Pool)
		if pool.Disable {
			return
		} else {
			return pool.Entries
		}
	}
	//cache miss
	poolJson, err := client.Get("/pools/" + name).Result()
	if log.IsDebug() {
		log.Debug("value from redis", "key", "/pools/"+name, "value", poolJson)
	}
	if err != nil || poolJson == "" {
		log.Warn("get pool from redis failed", "err", fmt.Sprintf("%v", err))
		return
	}
	pool := Pool{}
	json.Unmarshal([]byte(poolJson), &pool)
	poolCache.Put(name, pool)
	if pool.Disable {
		return
	} else {
		return pool.Entries
	}
}

func findEntriesByNameFromEtc(name string) (entries []Entry) {
	pool := Pools[name]
	if pool.Disable {
		return
	} else {
		entries = pool.Entries
	}
	log.Debug("merge pool host config", "name", name, "entries", entries)
	return
}

func GetEntriesByTargetsFromOrigin(targets []Target, randomize bool, fromRedis bool) (entries []Entry) {
	sort.Sort(sortTagets(targets))
	priorityFound := -1
	var availableTargets []Target
	for _, target := range targets {
		if priorityFound > -1 && target.Priority > priorityFound { //如果的优先级低于现有的优先级，无需运算
			log.Debug("ignore pool", "id", target.Pool, "priority", target.Priority, "current priority found", priorityFound)
			continue
		}
		log.Debug("find pool", "id", target.Pool, "priority", target.Priority)
		entriesFound := FindEntriesByName(target.Pool, false, fromRedis) //确认是否有可用entry
		if len(entriesFound) > 0 {                                       //如果有可用entry
			if target.Priority < priorityFound || priorityFound == -1 { //如果找到的新entry优先级更高，则原有的entry列表作废
				availableTargets = []Target{target}
				priorityFound = target.Priority
			} else { //同优先级的append
				availableTargets = append(availableTargets, target)
			}
		}
	}
	if randomize {
		availableTargets = rerHost.Randomize(availableTargets).([]Target)
	}
	for _, target := range availableTargets {
		entriesFound := FindEntriesByName(target.Pool, randomize, fromRedis)
		entries = append(entries, entriesFound...)
	}
	return
}

// modified by xuyingquan at date 2017.4.21
func GetEntriesByTargets(targets []Target, randomize bool, domain string, region Region, fromRedis bool) (entries []Entry) {
	sort.Sort(sortTagets(targets))
	priorityFound := -1
	var availableTargets []Target
	for _, target := range targets {
		if priorityFound > -1 && target.Priority > priorityFound { //如果的优先级低于现有的优先级，无需运算
			log.Debug("ignore pool", "id", target.Pool, "priority", target.Priority, "current priority found", priorityFound)
			continue
		}
		log.Debug("find pool", "id", target.Pool, "priority", target.Priority)
		entriesFound := FindEntriesByName(target.Pool, false, fromRedis) //确认是否有可用entry
		log.Debug("entries found", entriesFound)
		if len(entriesFound) > 0 { //如果有可用entry
			if target.Priority < priorityFound || priorityFound == -1 { //如果找到的新entry优先级更高，则原有的entry列表作废
				availableTargets = []Target{target}
				priorityFound = target.Priority
			} else { //同优先级的append
				availableTargets = append(availableTargets, target)
			}
		}
	}

	if len(availableTargets) == 0 { // 没有可用pool
		return
	}

	if len(availableTargets) == 1 { // 处理只有一个可用pool的情况
		log.Debug("only one pool")
		entriesFound := FindEntriesByName(availableTargets[0].Pool, randomize, fromRedis)
		entries = append(entries, entriesFound...)
		return
	}
	sort.Sort(SortTargetsByWeight(availableTargets)) // 按照权重进行排序

	var matrix *Matrix					// 生成的Matrix对象
	var historyRecord *AccessHistory	// 历史访问记录
	var access_key string				// 访问matrix key
	var access_matrix_index int 		// 上次访问Matrix索引
	var next_access_indexes []int		// 保存下次访问pools索引列表值
	var next_matrix_index int			// 保存下次访问matrix分解的索引

	pools, ips, weights := GetPoolsByTarget(availableTargets, fromRedis)
	matrix_str := fmt.Sprintf("%v-%v", ips, weights)
	matrix_key := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%v-%v", ips, weights))))
	log.Debug("GetEntriesByTargets", "matrix key", fmt.Sprintf("%v-%v", ips, weights), "md5", matrix_key)
	matrix = MatrixCache.GetMatrix(matrix_key)
	if matrix == nil { // 缓存没有
		matrix_data, err := util.SplitWeight(ips, weights)
		if err != nil {
			log.Warn("GetEntriesByTargets", "split error", err.Error())
			goto END
		}
		matrix = &Matrix{
			Data: matrix_data,
			IpnumList: ips,
			WeightList: weights,
		}
		MatrixCache.SetMatrix(matrix_key, matrix) // 设置缓存
	}

	access_key = fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s-%s-%s-%s", matrix_str, domain, region.Region, region.Isp))))
	log.Debug("GetEntriesByTargets", "access key", fmt.Sprintf("%s-%s-%s-%s", matrix_str, domain, region.Region, region.Isp), "md5", access_key)
	historyRecord = AccessHistoryCache.GetAccessHistory(access_key)
	if historyRecord == nil || len(historyRecord.AccessPoolIndex) != len(pools){ // 第一次访问
		historyRecord = &AccessHistory{
			AccessMatrixIndex: 0,
		}
		historyRecord.AccessPoolIndex = make([]int, 0)
		for i := 0; i < len(pools); i++ {
			historyRecord.AccessPoolIndex = append(historyRecord.AccessPoolIndex, 0)
		}
		AccessHistoryCache.SetAccessHistory(access_key, historyRecord) // 设置历史访问记录缓存
	}

	// 从pool里取出ips返回
	access_matrix_index = historyRecord.AccessMatrixIndex % len(matrix.Data)
	next_access_indexes = make([]int, 0)					// 保存下次访问pools索引列表值
	for i := 0; i < len(pools); i++ {
		j := historyRecord.AccessPoolIndex[i]		// 获取当前pool的索引变量

		if matrix.Data[access_matrix_index][i] == 0 { // 处理分解值为0的情况
			next_access_indexes = append(next_access_indexes, j)
			continue
		}

		count := 0
		for ; j < ips[i]; j=(j+1)%ips[i] {
			entries = append(entries, pools[i].Entries[j])
			count++
			if count == matrix.Data[access_matrix_index][i] { // 取够指定数目Entry
				break
			}
		}
		next_access_indexes = append(next_access_indexes, (j+1)%ips[i])
	}
	next_matrix_index = (access_matrix_index + 1) % len(matrix.Data)
	historyRecord = &AccessHistory{
		AccessMatrixIndex: next_matrix_index,
		AccessPoolIndex: next_access_indexes,
	}
	AccessHistoryCache.SetAccessHistory(access_key, historyRecord) // 更新访问历史记录

END:
	if len(entries) == 0 {
		for _, target := range availableTargets {
			entriesFound := FindEntriesByName(target.Pool, randomize, fromRedis)
			entries = append(entries, entriesFound...)
		}
	}

	// 随机打散ip
	entries = rerHost.Randomize(entries).([]Entry)
	return
}
