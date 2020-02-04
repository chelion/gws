package utils

import(
	"sync"
)

type ByteBufferPool struct{
	byteBufferPool	sync.Pool
	chunkSize 		int
	chunksPerAlloc int
}

func NewByteBufferPool(chunkSize,chunksPerAlloc int)(*ByteBufferPool){
	return &ByteBufferPool{chunkSize:chunkSize,chunksPerAlloc:chunksPerAlloc}
}

func (bbp *ByteBufferPool)Get()([]byte){
	v := bbp.byteBufferPool.Get()
	if v == nil {
		return make([]byte, bbp.chunkSize)
	}
	w := v.([]byte)
	return w
}

func (bbp *ByteBufferPool)Put(chunk []byte){
	bbp.byteBufferPool.Put(chunk)
}