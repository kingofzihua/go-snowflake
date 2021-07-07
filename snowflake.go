package snowflake

import (
	"errors"
	"time"
)

// 雪花ID部分的位长度
// *41位时间戳 + 10位工作机器ID + 12位序列号*
// 41位时间戳: 精确到毫秒级，41位的长度可以使用2^41=69年
// 10位工作机器ID: 最多支持部署2^10=1024个节点(一般是5位IDC+5位machine编号)
// 12位序列号: 支持每台服务器每毫秒产生 2^12=4096 个ID序号
const (
	TimestampLength = 41                     // 时间序列位
	MachineIDLength = 10                     // 机器ID
	SequenceLength  = 12                     // 序列号
	MaxSequence     = 1<<SequenceLength - 1  // 最大可用的序列号   2^12-1 =  4095
	MaxTimestamp    = 1<<TimestampLength - 1 // 最大可用时间 		2^41-1 = 2199023255551
	MaxMachineID    = 1<<MachineIDLength - 1 // 最大可用的机器ID 标识 2^10-1 =1023

	machineIDMoveLength = SequenceLength                   // 机器 ID 移动长度
	timestampMoveLength = MachineIDLength + SequenceLength // 时间序列移动长度
)

// SequenceResolver 序列号解决器
//
// When you want use the snowflake algorithm to generate unique ID, You must ensure: The sequence-number generated in the same millisecond of the same node is unique.
// Based on this, we create this interface provide following reslover:
//   AtomicResolver : base sync/atomic (默认).
type SequenceResolver func(ms int64) (uint16, error)

// default start time is 2008-11-10 23:00:00 UTC, why ? In the playground the time begins at 2009-11-10 23:00:00 UTC.
// It's can run on golang playground.
// default machineID is 0
// default resolver is AtomicResolver
var (
	resolver  SequenceResolver //生成器
	machineID = 0              //机器编号
	startTime = time.Date(2008, 11, 10, 23, 0, 0, 0, time.UTC) //开始时间
)

// ID 生成雪花ID 会忽略错误 如果需要错误，要使用 NextID
// 此函数是线程安全的。
func ID() uint64 {
	id, _ := NextID()
	return id
}

// NextID 使用 NextID 会生成雪花id 并返回错误
// 此函数是线程安全的。
func NextID() (uint64, error) {
	c := currentMillis()                  //获取当前毫秒数
	seqResolver := callSequenceResolver() // 获取序列号解析器
	seq, err := seqResolver(c)

	if err != nil {
		return 0, err
	}

	for seq >= MaxSequence { // 如果生成的编号大于最大编号，就等待下一秒
		c = waitForNextMillis(c)
		seq, err = seqResolver(c)
		if err != nil {
			return 0, err
		}
	}

	// 获取 当前日期和 开始时间的差值
	df := int(elapsedTime(c, startTime))
	if df < 0 || df > MaxTimestamp {// 如果超过了最大时间 需要用户调整开始时间
		return 0, errors.New("The maximum life cycle of the snowflake algorithm is 2^41-1(millis), please check starttime")
	}

	id := uint64((df << timestampMoveLength) | (machineID << machineIDMoveLength) | int(seq))
	return id, nil
}

// SetStartTime 设置雪花算法的开始时间
//
// 它会在以下情况下 panic:
//   s 是空
//   s 不能大于当前毫秒
//   当前毫秒数 - s > 2^41(69 年).
// This function is thread-unsafe, recommended you call him in the main function.
func SetStartTime(s time.Time) {
	s = s.UTC()

	if s.IsZero() {
		panic("The start time cannot be a zero value")
	}

	if s.After(time.Now()) {
		panic("The s cannot be greater than the current millisecond")
	}

	// 因为 s 必须在现在之前, 所以 `df` 不能小于 0。
	df := elapsedTime(currentMillis(), s)
	if df > MaxTimestamp {
		panic("The maximum life cycle of the snowflake algorithm is 69 years")
	}

	startTime = s
}

// SetMachineID specify the machine ID. It will panic when machineid > max limit for 2^10-1.
// This function is thread-unsafe, recommended you call him in the main function.
func SetMachineID(m uint16) {
	if m > MaxMachineID {
		panic("The machineid cannot be greater than 1023")
	}
	machineID = int(m)
}

// SetSequenceResolver 设置自定义 序列解析器
// 此函数是线程不安全的，建议您在主函数中调用他。
func SetSequenceResolver(seq SequenceResolver) {
	if seq != nil {
		resolver = seq
	}
}

// SID 雪花ID
type SID struct {
	Sequence  uint64 // 序列号
	MachineID uint64 // 机器ID
	Timestamp uint64 // 时间序列位
	ID        uint64
}

// GenerateTime 生成的时间, 返回一个 UTC time.
func (id *SID) GenerateTime() time.Time {
	ms := startTime.UTC().UnixNano()/1e6 + int64(id.Timestamp)

	return time.Unix(0, ms*int64(time.Millisecond)).UTC()
}

// ParseID 解析雪花ID 为一个 SID 的结构体.
func ParseID(id uint64) SID {
	timestamp := id >> timestampMoveLength
	sequence := id & MaxSequence
	machineID := (id & (MaxMachineID << SequenceLength)) >> SequenceLength

	return SID{
		ID:        id,
		Sequence:  sequence,
		MachineID: machineID,
		Timestamp: timestamp,
	}
}

//--------------------------------------------------------------------
// private function defined.
//--------------------------------------------------------------------

// 等待下一个毫秒
func waitForNextMillis(last int64) int64 {
	now := currentMillis()
	for now == last { //自旋等待下一毫秒
		now = currentMillis()
	}
	return now
}

// 获取序列号解析器
func callSequenceResolver() SequenceResolver {
	if resolver == nil {
		return AtomicResolver
	}

	return resolver
}

// 获取两个时间的差
func elapsedTime(nowMs int64, s time.Time) int64 {
	return nowMs - s.UTC().UnixNano()/1e6
}

// 获取当前毫秒数
func currentMillis() int64 {
	return time.Now().UTC().UnixNano() / 1e6
}
