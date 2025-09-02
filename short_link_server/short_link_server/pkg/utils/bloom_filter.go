package utils

import (
	"context"
	"hash/fnv"
	"math"

	"github.com/go-redis/redis/v8"
)

// BloomFilter 基于Redis的布隆过滤器实现
// 用于高效判断一个元素是否在集合中，具有零误报率（false negative rate）
// 但可能存在误判（false positive）
type BloomFilter struct {
	client      *redis.Client
	key         string        // Redis中的键名
	bitSize     uint          // 位图大小（位数）
	hashCount   uint          // 哈希函数数量
	expectedN   uint          // 预期元素数量
	falsePosPct float64       // 预期误判率
	ctx         context.Context
}

// NewBloomFilter 创建一个新的布隆过滤器
// expectedN: 预期元素数量
// falsePosPct: 预期误判率(如0.01表示1%)
func NewBloomFilter(client *redis.Client, key string, expectedN uint, falsePosPct float64) *BloomFilter {
	// 计算优化的位图大小(增加1.5倍以降低误判率): m = -n*ln(p)/(ln(2)^2) * 1.5，向上取整确保足够的位数
	bitSize := uint(math.Ceil(-float64(expectedN) * math.Log(falsePosPct) / (math.Log(2) * math.Log(2)) * 1.5))
	// 计算优化的哈希函数数量(增加1.5倍以降低误判率): k = (m/n)*ln(2) * 1.5，向上取整确保足够的哈希函数
	hashCount := uint(math.Ceil(float64(bitSize)/float64(expectedN) * math.Log(2) * 1.5))

	return &BloomFilter{
		client:      client,
		key:         key,
		bitSize:     bitSize,
	hashCount:   hashCount,
	expectedN:   expectedN,
	falsePosPct: falsePosPct,
	ctx:         context.Background(),
	}
}

// Add 添加元素到布隆过滤器
func (bf *BloomFilter) Add(element string) error {
	pipe := bf.client.Pipeline()

	for i := uint(0); i < bf.hashCount; i++ {
		idx := bf.getBitIndex(element, i)
		pipe.SetBit(bf.ctx, bf.key, int64(idx), 1)
	}

	_, err := pipe.Exec(bf.ctx)
	return err
}

// Contains 检查元素是否在布隆过滤器中
// 返回true表示可能存在，返回false表示一定不存在
func (bf *BloomFilter) Contains(element string) (bool, error) {
	pipe := bf.client.Pipeline()

	for i := uint(0); i < bf.hashCount; i++ {
		idx := bf.getBitIndex(element, i)
		pipe.GetBit(bf.ctx, bf.key, int64(idx))
	}

	results, err := pipe.Exec(bf.ctx)
	if err != nil {
		return false, err
	}

	for _, result := range results {
		bitValue := result.(*redis.IntCmd).Val()
		if bitValue == 0 {
			// 如果任何一个位为0，则元素一定不在集合中
			return false, nil
		}
	}

	// 所有位都为1，元素可能在集合中
	return true, nil
}

// getBitIndex 计算元素在指定哈希函数下的位图索引
// 使用双哈希技术生成多个哈希值: h_i(x) = h1(x) + i*h2(x)
func (bf *BloomFilter) getBitIndex(element string, hashIndex uint) uint {
	// 计算第一个哈希值
	h1 := fnv.New64a()
	h1.Write([]byte(element))
	v1 := h1.Sum64()

	// 计算第二个哈希值
	h2 := fnv.New64()
	h2.Write([]byte(element))
	v2 := h2.Sum64()

	// 生成第i个哈希值并取模
	combinedHash := v1 + uint64(hashIndex)*v2
	return uint(combinedHash % uint64(bf.bitSize))
}

// Stats 返回布隆过滤器的统计信息
func (bf *BloomFilter) Stats() map[string]interface{} {
	return map[string]interface{}{
		"key":          bf.key,
		"bitSize":      bf.bitSize,
		"hashCount":    bf.hashCount,
		"expectedN":    bf.expectedN,
		"falsePosPct": bf.falsePosPct,
	}
}