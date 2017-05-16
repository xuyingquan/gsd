package rander
import (
	"testing"
	"fmt"
)

type Item struct {
	value int
	Weight int
}

func TestRandomize(t *testing.T) {
	count := 60000
	array := []Item{
		Item{value: 1},
		Item{value: 2, Weight: 2},
		Item{value: 3, Weight: 3},
	}
	counter := []int{0, 0, 0}
	rander := New()
	for i := 0;i < count;i++ {
		newArray := (rander.Randomize(array)).([]Item)
		counter[newArray[0].value - 1] += 1
	}
	for _, c := range counter {
		fmt.Println(c)
	}
}
