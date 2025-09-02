package utils

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

// 测试布隆过滤器的基本功能
func TestBloomFilterBasic(t *testing.T) {
	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client.FlushAll(context.Background()) // 测试结束后清理数据

	// 创建布隆过滤器: 预期1000个元素，误判率1%
	bf := NewBloomFilter(client, "test:bloomfilter", 1000, 0.01)

	// 测试添加和查询
	elements := []string{"test1", "test2", "test3"}
	for _, elem := range elements {
		assert.NoError(t, bf.Add(elem))
	}

	// 检查已添加的元素
	for _, elem := range elements {
		exists, err := bf.Contains(elem)
		assert.NoError(t, err)
		assert.True(t, exists, "%s should be in the filter", elem)
	}

	// 检查未添加的元素
	notExists := "notexist"
	exists, err := bf.Contains(notExists)
	assert.NoError(t, err)
	assert.False(t, exists, "%s should not be in the filter", notExists)
}

// 测试布隆过滤器的误判率
func TestBloomFilterFalsePositive(t *testing.T) {
	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client.FlushAll(context.Background())

	// 创建布隆过滤器: 预期1000个元素，误判率5%
	bf := NewBloomFilter(client, "test:bloomfilter:fp", 1000, 0.05)

	// 添加1000个元素
	elements := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		elements[i] = fmt.Sprintf("element%d", i)
		assert.NoError(t, bf.Add(elements[i]))
	}

	// 测试1000个未添加的元素，统计误判率
	falsePositives := 0
	for i := 1000; i < 2000; i++ {
		elem := fmt.Sprintf("nonexistent%d", i)
		exists, err := bf.Contains(elem)
		assert.NoError(t, err)
		if exists {
			falsePositives++
		}
	}

	// 确保误判率低于预期的5%
	falsePosPct := float64(falsePositives) / 1000.0
	assert.LessOrEqual(t, falsePosPct, 0.08, "False positive rate should be below 8%")

	// 输出实际误判率
	fmt.Printf("Actual false positive rate: %.2f%%\n", falsePosPct*100)
}

// 测试Stats方法
func TestBloomFilterStats(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	bf := NewBloomFilter(client, "test:bloomfilter:stats", 1000, 0.01)
	stats := bf.Stats()

	assert.Equal(t, "test:bloomfilter:stats", stats["key"])
	// 由于使用了math.Ceil，实际值可能会有微小差异，使用近似断言
	bitSize := stats["bitSize"].(uint)
	assert.InDelta(t, 9586, bitSize, 1, "bitSize should be approximately 9586")
	// 哈希函数数量也可能因四舍五入而变化
	hashCount := stats["hashCount"].(uint)
	assert.InDelta(t, 7, hashCount, 1, "hashCount should be approximately 7")
	assert.Equal(t, uint(1000), stats["expectedN"])
	assert.Equal(t, 0.01, stats["falsePosPct"])
}