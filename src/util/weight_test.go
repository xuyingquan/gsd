package util

import (
	"testing"
	"fmt"
)

func TestSplitWeight(t *testing.T) {
	ipnumList := []int{2, 3, 1}
	weightList := []int{4, 2, 1}
	fmt.Printf("测试数据: ipnumList=%v, weightList=%v\n", ipnumList, weightList)
	matrix, _ := SplitWeight(ipnumList, weightList)
	fmt.Printf("测试结果: %v\n", matrix)

	ipnumList = []int{2, 2, 2}
	weightList = []int{9, 9, 1}
	fmt.Printf("测试数据: ipnumList=%v, weightList=%v\n", ipnumList, weightList)
	matrix, _ = SplitWeight(ipnumList, weightList)
	fmt.Printf("测试结果: %v\n", matrix)

	ipnumList = []int{2, 2}
	weightList = []int{9, 1}
	fmt.Printf("测试数据: ipnumList=%v, weightList=%v\n", ipnumList, weightList)
	matrix, _ = SplitWeight(ipnumList, weightList)
	fmt.Printf("测试结果: %v\n", matrix)

	ipnumList = []int{1, 1}
	weightList = []int{100000, 1}
	fmt.Printf("测试数据: ipnumList=%v, weightList=%v\n", ipnumList, weightList)
	matrix, _ = SplitWeight(ipnumList, weightList)
	fmt.Printf("测试结果: %v\n", matrix)

}

func BenchmarkSplitWeight(b *testing.B) {
	ipnumList := []int{2, 3, 1}
	weightList := []int{4, 2, 1}
	for i := 0; i < b.N; i++ {
		SplitWeight(ipnumList, weightList)
	}
}
