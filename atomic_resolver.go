package snowflake

import (
	"sync/atomic"
)

var lastTime int64 //最后生成时间
var lastSeq uint32 //最后生成序列

// AtomicResolver define as atomic sequence resolver, base on standard sync/atomic.
func AtomicResolver(ms int64) (uint16, error) {
	var last int64
	var seq, localSeq uint32
	for {
		last = atomic.LoadInt64(&lastTime)
		localSeq = atomic.LoadUint32(&lastSeq)

		if last > ms { // 当前毫秒数小于最后生成的
			return MaxSequence, nil
		}

		if last == ms {
			seq = MaxSequence & (localSeq + 1)
			if seq == 0 {
				return MaxSequence, nil
			}
		}

		if atomic.CompareAndSwapInt64(&lastTime, last, ms) && atomic.CompareAndSwapUint32(&lastSeq, localSeq, seq) {
			return uint16(seq), nil
		}
	}
}
