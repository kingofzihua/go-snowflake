package snowflake_test

import (
	"fmt"
	"github.com/kingofzihua/go-snowflake"
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

	<- ch
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
