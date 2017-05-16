package rander

import (
	"math/rand"
	"reflect"
	"sync"
	"time"
)

type Rander struct {
	r    *rand.Rand
	lock sync.Mutex
}

func New() (rander *Rander) {
	rander = &Rander{r: rand.New(rand.NewSource(time.Now().UnixNano()))}
	return
}

func (rander *Rander) Random(min int, max int) int {
	return min + rander.r.Intn(max-min)
}

func (rander *Rander) Randomize(array interface{}) (result interface{}) {
	if array == nil {
		return
	}
	v := reflect.ValueOf(array)
	if v.Kind() != reflect.Array && v.Kind() != reflect.Slice || v.Len() < 1 {
		return array
	}
	length := v.Len()
	weights := make([]int64, length)
	weightSum := int64(0)
	resultV := reflect.MakeSlice(v.Type(), 0, length)
	for i := 0; i < length; i++ {
		item := v.Index(i)
		weight := item.FieldByName("Weight").Int()
		if weight == 0 { //default weight is 1
			weight = 1
		}
		weightSum += weight
		weights[i] = weight
	}
	rander.lock.Lock()
	r := rander.r.Int63n(weightSum)
	rander.lock.Unlock()
	start := 0
	weightPoint := int64(0)
	for i, weight := range weights {
		if r >= weightPoint && r < weightPoint+weight {
			start = i
			break
		}
		weightPoint += weight
	}
	for i := start; i < length+start; i++ {
		resultV = reflect.Append(resultV, v.Index(i%length))
	}
	result = resultV.Interface()
	return
}
