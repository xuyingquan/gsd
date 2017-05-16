package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"rule"
	"runtime"
)

const (
	Version = "5.2.0"
)

func findRegion(ipdb *rule.IPDB, ip string) {
	parsedIp := net.ParseIP(ip)
	if parsedIp == nil {
		fmt.Println("错误的IP地址@")
		os.Exit(1)
	}
	region := ipdb.FindRegionByIp(int64(binary.BigEndian.Uint32(parsedIp.To4()[:])))
	fmt.Printf("%s@%s\n", region.Region, region.Isp)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: ipfrom -db ipdb.db ip\n")
		fmt.Fprintf(os.Stderr, "command line switches:\n")
		flag.PrintDefaults()
	}
	dbPath := flag.String("db", "./ipdb.db", "ipdb path")
	version := flag.Bool("v", false, "show version and exit")
	ipToInt := flag.Bool("toint", false, "convert ip to int")
	pipeMode := flag.Bool("p", false, "pipe mode")
	help := flag.Bool("h", false, "show help and exit")
	flag.Parse()

	if *version {
		fmt.Printf("Ip Come From tool v%s build by %s\n", Version, runtime.Version())
		os.Exit(0)
	}

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		os.Exit(1)
	}
	ipdb := &rule.IPDB{}
	ipdb.LoadIPDB(*dbPath)
	for _, ip := range flag.Args() {
		if *ipToInt {
			pip := net.ParseIP(ip)
			fmt.Printf("%d\n", binary.BigEndian.Uint32(pip.To4()[:]))
			return
		}
		if *pipeMode {
			fmt.Println("pipe mode")
			pr := &io.PipeReader{}
			bpr := bufio.NewReader(pr)
			iip, _, err := bpr.ReadLine()
			if err == nil {
				fmt.Println(string(iip))
				findRegion(ipdb, string(iip))
			} else {
				fmt.Println(err)
			}
		} else {
			findRegion(ipdb, ip)
		}
	}

}
