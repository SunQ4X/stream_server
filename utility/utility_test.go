package utility

import (
	"fmt"
	"testing"
	"time"
)

var (
	cache = NewCache()
)

type TestCounter struct {
	*ReferenceCounter
}

func NewTestCounter(key string) *TestCounter {
	return &TestCounter{ReferenceCounter: NewReferenceCounter(key, time.Second)}
}

func (counter *TestCounter) Erase() {
}

func Test(t *testing.T) {
	timer := time.NewTicker(time.Millisecond * 50)
	go func() {
		for {
			<-timer.C
			func() {
				cache.mutex.Lock()
				defer cache.mutex.Unlock()

				fmt.Println(cache.elements)
			}()
		}
	}()

	counter := NewTestCounter("123")
	counter.AddReference()

	cache.Add(counter.Key(), counter)

	time.Sleep(time.Second)

	counter2 := NewTestCounter("abc")
	counter2.AddReference()
	cache.Add(counter2.Key(), counter2)

	time.Sleep(time.Second)

	counter.DelReference()

	time.Sleep(time.Second)

	counter2.DelReference()

	time.Sleep(time.Second)
}
