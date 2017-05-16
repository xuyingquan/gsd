package dnss

import (
	"fmt"
	//	"math/rand"
	"net"
	"rander"
	"reflect"
	"rule"
	"runtime/debug"
	"strings"
	"time"

	"github.com/mgutz/logxi/v1"
	"github.com/miekg/dns"
)

type Handler struct{}

type accesslog struct {
	ReqTime    time.Time //请求开始时间
	RespTime   time.Time //请求响应时间
	Domain     string    //请求的域名
	RemoteIp   string
	IsEDns     bool     //是否EDNS
	RecordType string   //返回记录类型，A或CNAME
	IsFound    bool     //是否找到结果
	IpList     []string //返回的ip列表
	Labels     map[string]string
}

var rand *rander.Rander = rander.New()

var logFormat = "$ReqTime|warn|$RemoteIp|$Labels.channel|$Labels.region|$Domain|$IsEDns|$RecordType|$IsFound|$IpList"

// Added by xuyingquan at 2017.4.20
// 产生一个区间随机数
//func randInt(min int, max int) int {
//	rand.Seed(time.Now().UnixNano())
//	return min + rand.Intn(max-min)
//}

func (h *Handler) handle(w dns.ResponseWriter, req *dns.Msg) {
	defer func() {
		if x := recover(); x != nil {
			fmt.Println("panic:", x)
			debug.PrintStack()
			log.Error("panic", "request domain", req.Question[0].Name, "remote addr", w.RemoteAddr().String())
		}
	}()
	aLog := accesslog{IpList: []string{}}
	aLog.ReqTime = time.Now()
	qtype := req.Question[0].Qtype
	name := req.Question[0].Name
	aLog.Domain = name[:len(name)-1]
	log.Debug("host request", "req", fmt.Sprintf("%v", req), "qtype", qtype)
	msg := new(dns.Msg)
	msg.SetReply(req)
	msg.Authoritative = true
	msg.RecursionAvailable = false // We're a nameserver, no recursion available
	zone := rule.FindZone(name, false)
	if qtype == dns.TypeA || qtype == dns.TypeAAAA || qtype == dns.TypeCNAME {
		if qtype == dns.TypeAAAA {
			aLog.RecordType = "AAAA"
		}
		ip, isEDns := getRemoteIp(w, req)
		aLog.RemoteIp = ip.String()
		aLog.IsEDns = isEDns

		records, nsRecords, labels := rule.FindRecords(strings.ToLower(name), ip, false)
		aLog.Labels = labels
		log.Debug("find records", "records", fmt.Sprintf("%v", records))
		log.Info("request", "name", name, "qtype", qtype, "records", records)
		if records != nil && len(*records) > 0 {
			for _, record := range *records {
				log.Debug("response record", "record", fmt.Sprintf("%v", record))
				recordType := dns.TypeA
				switch record.RecordType {
				case rule.TypeA:
					if qtype == dns.TypeA {
						aLog.RecordType = "A"
						recordType = dns.TypeA
					}
				case rule.TypeCNAME:
					aLog.RecordType = "CNAME"
					recordType = dns.TypeCNAME
				}
				header := dns.RR_Header{Name: name, Class: dns.ClassINET, Rrtype: recordType, Ttl: record.Ttl}
				switch record.RecordType {
				case rule.TypeA:
					if qtype == dns.TypeA {
						ip := record.Ip
						rr := &dns.A{header, ip}
						msg.Answer = append(msg.Answer, rr)
						aLog.IpList = append(aLog.IpList, ip.String())
					}
				case rule.TypeCNAME:
					rr := &dns.CNAME{header, record.CName}
					msg.Answer = append(msg.Answer, rr)
					aLog.IpList = append(aLog.IpList, record.CName)
				}
				aLog.IsFound = true
			}
			if len(msg.Answer) > 0 {
				if len(nsRecords) > 0 {
					msg.AuthenticatedData = true
					for _, nsRecord := range nsRecords {
						header := dns.RR_Header{Name: dns.Fqdn(nsRecord.Name), Class: dns.ClassINET, Rrtype: dns.TypeNS, Ttl: 86400}
						rr := &dns.NS{header, dns.Fqdn(nsRecord.Ns)}
						msg.Ns = append(msg.Ns, rr)
					}
				} else {
					if zone != nil {
						if len(zone.Ns) > 0 {
							msg.Ns = []dns.RR{}
							for _, nsr := range zone.Ns {
								header := dns.RR_Header{Name: dns.Fqdn(zone.Origin), Class: dns.ClassINET, Rrtype: dns.TypeNS, Ttl: 86400}
								rr := &dns.NS{header, dns.Fqdn(nsr.Name)}
								msg.Ns = append(msg.Ns, rr)
							}
						}
					}
				}
			}
		}
	} else if qtype == dns.TypeSOA {
		aLog.RecordType = "SOA"
		//不支持AAAA，返回空，SOA不需要添加，值为空时会自动添加
	} else if qtype == dns.TypeNS {
		aLog.RecordType = "NS"
		zone := rule.FindZone(name, true)
		if zone != nil {
			if len(zone.Ns) > 0 {
				msg.Ns = []dns.RR{}
				nsList := (rand.Randomize(zone.Ns)).([]rule.NsInZone)
				for _, nsr := range nsList {
					header := dns.RR_Header{Name: dns.Fqdn(zone.Origin), Class: dns.ClassINET, Rrtype: dns.TypeNS, Ttl: 86400}
					rr := &dns.NS{header, dns.Fqdn(nsr.Name)}
					msg.Ns = append(msg.Ns, rr)
					aHeader := dns.RR_Header{Name: dns.Fqdn(nsr.Name), Class: dns.ClassINET, Rrtype: dns.TypeA, Ttl: 7200}
					arr := &dns.A{aHeader, net.ParseIP(nsr.Ip)}
					msg.Extra = append(msg.Extra, arr)
					aLog.IpList = append(aLog.IpList, nsr.Name)
				}
			}
		}
	} else {
		msg.SetRcode(req, dns.RcodeNameError)
		msg.Authoritative = true
		msg.RecursionAvailable = false

		// Add a useful TXT record
		header := dns.RR_Header{Name: req.Question[0].Name,
			Class:  dns.ClassINET,
			Rrtype: dns.TypeTXT}
		msg.Ns = []dns.RR{&dns.TXT{header, []string{"Record type not supported"}}}
	}

	if len(msg.Answer) == 0 && len(msg.Ns) == 0 {
		if zone != nil {
			rr, err := dns.NewRR(zone.Soa)
			if err != nil {
				log.Error("SOA Error", "err", err)
			} else {
				msg.Ns = []dns.RR{rr}
			}
		}
	}
	aLog.RespTime = time.Now()
	err := w.WriteMsg(msg)
	if err != nil {
		log.Error("response error:", "err", fmt.Sprintf("%v", err))
	}
	printAccessLog(aLog)
	return
}

func getRemoteIp(w dns.ResponseWriter, req *dns.Msg) (ip net.IP, isEDns bool) { // EDNS or real IP
	realIp, _, _ := net.SplitHostPort(w.RemoteAddr().String())

	for _, extra := range req.Extra { //查找EDNS记录，如果有，以EDNS为准

		switch extra.(type) {
		case *dns.OPT:
			for _, o := range extra.(*dns.OPT).Option {
				switch e := o.(type) {
				case *dns.EDNS0_NSID:
				// do stuff with e.Nsid
				case *dns.EDNS0_SUBNET:
					if e.Address != nil && e.Family == 1 {
						ip = e.Address
						isEDns = true
					}
				}
			}
		}
	}

	if len(ip) == 0 { // no edns subnet
		isEDns = false
		ip = net.ParseIP(realIp)
	}

	log.Info("query by remote ip", "ip", ip.To4()[:].String())
	return
}

func ListenAndServe(addr string, confDir string) {
	rule.LoadConf(confDir)

	handler := &Handler{}
	udpHandler := dns.NewServeMux()
	tcpHandler := dns.NewServeMux()
	udpHandler.HandleFunc(".", handler.handle)
	tcpHandler.HandleFunc(".", handler.handle)

	tcpServer := &dns.Server{
		Addr:    addr,
		Net:     "tcp",
		Handler: tcpHandler,
	}

	udpServer := &dns.Server{
		Addr:    addr,
		Net:     "udp",
		Handler: udpHandler,
		UDPSize: 65535,
	}

	go func() {
		log.Info("GSD Server Start!Listen on tcp", "addr", addr)
		err := tcpServer.ListenAndServe()
		if err != nil {
			log.Error("tcp server error", "err", err)
		}
	}()
	log.Info("Listen on udp", "addr", addr)
	err := udpServer.ListenAndServe()
	if err != nil {
		log.Error("udp server error", "err", err)
	}

	//go rule.CleanCache() // 清理缓存，每个24小时清理一次
}

func Reload() {
	rule.Reload()
}

func printAccessLog(aLog accesslog) {
	columns := strings.Split(logFormat, "|")
	output := ""
	value := reflect.ValueOf(aLog)
	for _, c := range columns {
		if c[0] == '$' {
			field := string(c[1:])
			fm := strings.Split(field, ".") //$aaa.bbb
			fieldValue := reflect.Value{}
			if len(fm) == 2 { //采用map方式取值
				fieldMap := value.FieldByName(fm[0])
				if fieldMap.IsValid() && fieldMap.Kind() == reflect.Map {
					fieldValue = fieldMap.MapIndex(reflect.ValueOf(fm[1]))
				}
			} else {
				fieldValue = value.FieldByName(field)
			}

			if fieldValue.IsValid() {
				if fieldValue.Type().String() == "time.Time" {
					output += fieldValue.MethodByName("Format").Call([]reflect.Value{reflect.ValueOf("2006-01-02 15:04:05")})[0].String()
				} else if fieldValue.Type().String() == "bool" {
					if fieldValue.Bool() {
						output += "1"
					} else {
						output += "0"
					}
				} else if fieldValue.Type().String() == "[]string" {
					list := ""
					for i := 0; i < fieldValue.Len(); i++ {
						list += fieldValue.Index(i).String() + ","
					}
					if len(list) > 0 {
						output += list[:len(list)-1]
					}
				} else {
					output += fieldValue.String()
				}
			}
		} else {
			output += c
		}
		output += "|"
	}
	fmt.Println("ACCESSLOG:", output[:len(output)-1])
}
