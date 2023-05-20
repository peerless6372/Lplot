package utils

/*
	缓存池，避免重复申请内存，减轻gc压力
	一个sync.Pool对象就是一组临时对象的集合。Pool是协程安全的
  	注意事项：获取资源Get后，必须执行Put放回操作
*/
import (
	"bytes"
	"sync"
)

type BufferPool struct {
	sync.Pool
}

// 声明一块buffer pool大小
func NewBufferPool(bufferSize int) (bufferpool *BufferPool) {
	return &BufferPool{
		sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 0, bufferSize))
			},
		},
	}
}

// 获取一块buffer写入
func (bufferpool *BufferPool) Get() *bytes.Buffer {
	return bufferpool.Pool.Get().(*bytes.Buffer)
}

// 放回buffer重用
func (bufferpool *BufferPool) Put(b *bytes.Buffer) {
	b.Reset()
	bufferpool.Pool.Put(b)
}
