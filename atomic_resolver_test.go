package snowflake_test

import (
	"fmt"
	"github.com/kingofzihua/go-snowflake"
	"sync/atomic"
	"testing"
	"time"
)

func TestAtomicResolver(t *testing.T) {
	id, _ := snowflake.AtomicResolver(1)

	if id != 0 {
		t.Error("Sequence should be equal 0")
	}
}

func TestAtomicResolver2(t *testing.T) {
	var ch = make(chan bool)

	go func() {
		<-time.After(time.Second)
		ch <- true
	}()

	go func() {
		for {
			id, _ := snowflake.AtomicResolver(1)
			fmt.Println(id)
		}
	}()

	<-ch
}
func TestAtomicResolver3(t *testing.T) {
	var id uint16
	id, _ = snowflake.AtomicResolver(1)
	fmt.Println(id)
	id, _ = snowflake.AtomicResolver(1)
	fmt.Println(id)
	id, _ = snowflake.AtomicResolver(2)
	fmt.Println(id)
	id, _ = snowflake.AtomicResolver(2)
	fmt.Println(id)
	id, _ = snowflake.AtomicResolver(2)
	fmt.Println(id)
}

func BenchmarkCombinationParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = snowflake.AtomicResolver(1)
		}
	})
}

func BenchmarkAtomicResolver(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = snowflake.AtomicResolver(1)
	}
}

func TestCAS(t *testing.T) {

	var lastTime int64 //最后生成时间

	var last = atomic.LoadInt64(&lastTime)

	for i := 0; i < 10; i++ {
		res := atomic.CompareAndSwapInt64(&lastTime, last, 1)

		fmt.Println(res)
	}

}
