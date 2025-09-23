package user

// 位运算工具函数，用于处理月签到和奖励领取的位图存储

// SetBit 设置指定位为1
func SetBit(value int32, bit int32) int32 {
	if bit < 1 || bit > 31 {
		return value // 超出范围，不修改
	}
	return value | (1 << (bit - 1)) // 位从1开始，所以减1
}

// ClearBit 设置指定位为0
func ClearBit(value int32, bit int32) int32 {
	if bit < 1 || bit > 31 {
		return value // 超出范围，不修改
	}
	return value &^ (1 << (bit - 1)) // 位从1开始，所以减1
}

// GetBit 获取指定位的值（0或1）
func GetBit(value int32, bit int32) bool {
	if bit < 1 || bit > 31 {
		return false // 超出范围，返回false
	}
	return (value & (1 << (bit - 1))) != 0 // 位从1开始，所以减1
}

// GetSetBits 获取所有设置为1的位（返回日期列表）
func GetSetBits(value int32) []int32 {
	var bits []int32
	for i := int32(1); i <= 31; i++ {
		if GetBit(value, i) {
			bits = append(bits, i)
		}
	}
	return bits
}

// CountBits 计算设置为1的位数
func CountBits(value int32) int32 {
	count := int32(0)
	for value > 0 {
		if value&1 == 1 {
			count++
		}
		value >>= 1
	}
	return count
}

// HasBit 检查指定位是否已设置
func HasBit(value int32, bit int32) bool {
	return GetBit(value, bit)
}

// IsEmpty 检查位图是否为空（所有位都为0）
func IsEmpty(value int32) bool {
	return value == 0
}

// IsFull 检查位图是否已满（所有位都为1）
func IsFull(value int32) bool {
	return value == 0x7FFFFFFF // 31位全为1
}
