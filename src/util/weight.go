package util

import (
	"errors"
	"github.com/mgutz/logxi/v1"
	"fmt"
)

func SplitWeight(ipnumList []int, weightList []int) (matrix [][]int, err error) {
	if len(ipnumList) != len(weightList) {
		return nil, errors.New("ipnumList not equal weightList")
	}

	result_matrix := make([][]int, 0) 			// 分解矩阵

	var SplitTries int = 0 						//  权重拆分次数
	for index, weight := range weightList {
		weights := make([]int, 0) 				// 分解后的权重列表

		if index == 0 { // 分解第一个pool权重
			for {
				if weight <= ipnumList[index] { // 不用再拆分
					SplitTries++
					weights = append(weights, weight)
					break
				} else {
					weights = append(weights, ipnumList[index])
					weight -= ipnumList[index]
				}
				SplitTries++
			}
			result_matrix = append(result_matrix, weights)
		} else { // 分解其他pool权重
			for i := 0; i < SplitTries; i++ {
				if i == SplitTries-1 && weight > 1 { // 最后一次分解且未分解权重值较大
					weights = append(weights, weight)
					break
				}
				if weight-(SplitTries-i)+1 >= ipnumList[index] { // 保证余下分解有1
					weights = append(weights, ipnumList[index])
					weight -= ipnumList[index]
					continue
				} else {
					if weight == 0 {
						weights = append(weights, 0)
						continue
					}
					weights = append(weights, 1)
					weight -= 1
				}
			}
			result_matrix = append(result_matrix, weights)
		}
	}

	// 倒置矩阵
	reverse_matrix := make([][]int, 0)
	for i := 0; i < SplitTries; i++ {
		reverse_matrix = append(reverse_matrix, make([]int, 0))
	}
	for i := range result_matrix {
		for j := range result_matrix[i] {
			reverse_matrix[j] = append(reverse_matrix[j], result_matrix[i][j])
		}
	}
	log.Debug("SplitWeight", "split matrix", reverse_matrix)

	// 去除单pool的分解
	matrix = make([][]int, 0)
	for i := range reverse_matrix {
		count := 0
		for j := range reverse_matrix[i] {
			if reverse_matrix[i][j] > 0 {
				count++
			}
		}
		if count > 1 {
			matrix = append(matrix, reverse_matrix[i])
		}
	}

	if len(matrix) == 0 {
		err_str := fmt.Sprintf("can't split ipnumList=%v, weightList=%v", ipnumList, weightList)
		return nil, errors.New(err_str)
	}
	log.Debug("SplitWeight", "available matrix", matrix)
	return
}
