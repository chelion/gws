package utils

import(
	"fmt"
	"sync"
	"syscall"
	"unsafe"
)

type ByteBufferPool struct{
	chunksPerAlloc  int
	chunkSize 		int
	freeChunks      []interface{}
	freeChunksLock *sync.Mutex
}

func NewByteBufferPool(chunkSize,chunksPerAlloc int)(*ByteBufferPool){
	return &ByteBufferPool{chunkSize:chunkSize,chunksPerAlloc:chunksPerAlloc,
		freeChunks:make([]interface{},0),freeChunksLock:new(sync.Mutex)}
}

func (bbp *ByteBufferPool)Get()([]byte){
	s := (bbp.chunkSize+1023)/1024
	bbp.freeChunksLock.Lock()
	if len(bbp.freeChunks) == 0 {
		data, err := syscall.Mmap(-1, 0, bbp.chunkSize*bbp.chunksPerAlloc, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_ANON|syscall.MAP_PRIVATE)
		if err != nil {
			panic(fmt.Errorf("cannot allocate %d bytes via mmap: %s", bbp.chunkSize*bbp.chunksPerAlloc, err))
		}
		for len(data) > 0 {
			switch(s){
				case 1:{
					p := (*[1024*1]byte)(unsafe.Pointer(&data[0]))
					bbp.freeChunks = append(bbp.freeChunks, p)
					data = data[1024*1:]
				}
				case 2:{
					p := (*[1024*2]byte)(unsafe.Pointer(&data[0]))
					bbp.freeChunks = append(bbp.freeChunks, p)
					data = data[1024*2:]
				}
				case 3:{
					p := (*[1024*3]byte)(unsafe.Pointer(&data[0]))
					bbp.freeChunks = append(bbp.freeChunks, p)
					data = data[1024*3:]
				}
				case 4:{
					p := (*[1024*4]byte)(unsafe.Pointer(&data[0]))
					bbp.freeChunks = append(bbp.freeChunks, p)
					data = data[1024*4:]
				}
				case 5:{
					p := (*[1024*5]byte)(unsafe.Pointer(&data[0]))
					bbp.freeChunks = append(bbp.freeChunks, p)
					data = data[1024*5:]
				}
				case 6:{
					p := (*[1024*6]byte)(unsafe.Pointer(&data[0]))
					bbp.freeChunks = append(bbp.freeChunks, p)
					data = data[1024*6:]
				}
				case 7:{
					p := (*[1024*7]byte)(unsafe.Pointer(&data[0]))
					bbp.freeChunks = append(bbp.freeChunks, p)
					data = data[1024*7:]
				}
				case 8:{
					p := (*[1024*8]byte)(unsafe.Pointer(&data[0]))
					bbp.freeChunks = append(bbp.freeChunks, p)
					data = data[1024*8:]
				}
				case 9:{
					p := (*[1024*9]byte)(unsafe.Pointer(&data[0]))
					bbp.freeChunks = append(bbp.freeChunks, p)
					data = data[1024*9:]
				}
				case 10:{
					p := (*[1024*10]byte)(unsafe.Pointer(&data[0]))
					bbp.freeChunks = append(bbp.freeChunks, p)
					data = data[1024*10:]
				}
				case 11:{
					p := (*[1024*11]byte)(unsafe.Pointer(&data[0]))
					bbp.freeChunks = append(bbp.freeChunks, p)
					data = data[1024*11:]
				}
				case 12:{
					p := (*[1024*12]byte)(unsafe.Pointer(&data[0]))
					bbp.freeChunks = append(bbp.freeChunks, p)
					data = data[1024*12:]
				}
				case 13:{
					p := (*[1024*13]byte)(unsafe.Pointer(&data[0]))
					bbp.freeChunks = append(bbp.freeChunks, p)
					data = data[1024*13:]
				}
				case 14:{
					p := (*[1024*14]byte)(unsafe.Pointer(&data[0]))
					bbp.freeChunks = append(bbp.freeChunks, p)
					data = data[1024*14:]
				}
				case 15:{
					p := (*[1024*15]byte)(unsafe.Pointer(&data[0]))
					bbp.freeChunks = append(bbp.freeChunks, p)
					data = data[1024*15:]
				}
				default:
				fallthrough
				case 16:{
					p := (*[1024*16]byte)(unsafe.Pointer(&data[0]))
					bbp.freeChunks = append(bbp.freeChunks, p)
					data = data[1024*16:]
				}
			}
		}
	}
	switch(s){
		case 1:{
			n := len(bbp.freeChunks) - 1
			r := bbp.freeChunks[n].(*[1024*1]byte)
			bbp.freeChunks[n] = nil
			bbp.freeChunks = bbp.freeChunks[:n]
			bbp.freeChunksLock.Unlock()
			return r[:]
		}
		case 2:{
			n := len(bbp.freeChunks) - 1
			r := bbp.freeChunks[n].(*[1024*2]byte)
			bbp.freeChunks[n] = nil
			bbp.freeChunks = bbp.freeChunks[:n]
			bbp.freeChunksLock.Unlock()
			return r[:]
		}
		case 3:{
			n := len(bbp.freeChunks) - 1
			r := bbp.freeChunks[n].(*[1024*3]byte)
			bbp.freeChunks[n] = nil
			bbp.freeChunks = bbp.freeChunks[:n]
			bbp.freeChunksLock.Unlock()
			return r[:]
		}
		case 4:{
			n := len(bbp.freeChunks) - 1
			r := bbp.freeChunks[n].(*[1024*4]byte)
			bbp.freeChunks[n] = nil
			bbp.freeChunks = bbp.freeChunks[:n]
			bbp.freeChunksLock.Unlock()
			return r[:]
		}
		case 5:{
			n := len(bbp.freeChunks) - 1
			r := bbp.freeChunks[n].(*[1024*5]byte)
			bbp.freeChunks[n] = nil
			bbp.freeChunks = bbp.freeChunks[:n]
			bbp.freeChunksLock.Unlock()
			return r[:]
		}
		case 6:{
			n := len(bbp.freeChunks) - 1
			r := bbp.freeChunks[n].(*[1024*6]byte)
			bbp.freeChunks[n] = nil
			bbp.freeChunks = bbp.freeChunks[:n]
			bbp.freeChunksLock.Unlock()
			return r[:]
		}
		case 7:{
			n := len(bbp.freeChunks) - 1
			r := bbp.freeChunks[n].(*[1024*7]byte)
			bbp.freeChunks[n] = nil
			bbp.freeChunks = bbp.freeChunks[:n]
			bbp.freeChunksLock.Unlock()
			return r[:]
		}
		case 8:{
			n := len(bbp.freeChunks) - 1
			r := bbp.freeChunks[n].(*[1024*8]byte)
			bbp.freeChunks[n] = nil
			bbp.freeChunks = bbp.freeChunks[:n]
			bbp.freeChunksLock.Unlock()
			return r[:]
		}
		case 9:{
			n := len(bbp.freeChunks) - 1
			r := bbp.freeChunks[n].(*[1024*9]byte)
			bbp.freeChunks[n] = nil
			bbp.freeChunks = bbp.freeChunks[:n]
			bbp.freeChunksLock.Unlock()
			return r[:]
		}
		case 10:{
			n := len(bbp.freeChunks) - 1
			r := bbp.freeChunks[n].(*[1024*10]byte)
			bbp.freeChunks[n] = nil
			bbp.freeChunks = bbp.freeChunks[:n]
			bbp.freeChunksLock.Unlock()
			return r[:]
		}
		case 11:{
			n := len(bbp.freeChunks) - 1
			r := bbp.freeChunks[n].(*[1024*11]byte)
			bbp.freeChunks[n] = nil
			bbp.freeChunks = bbp.freeChunks[:n]
			bbp.freeChunksLock.Unlock()
			return r[:]
		}
		case 12:{
			n := len(bbp.freeChunks) - 1
			r := bbp.freeChunks[n].(*[1024*12]byte)
			bbp.freeChunks[n] = nil
			bbp.freeChunks = bbp.freeChunks[:n]
			bbp.freeChunksLock.Unlock()
			return r[:]
		}
		case 13:{
			n := len(bbp.freeChunks) - 1
			r := bbp.freeChunks[n].(*[1024*13]byte)
			bbp.freeChunks[n] = nil
			bbp.freeChunks = bbp.freeChunks[:n]
			bbp.freeChunksLock.Unlock()
			return r[:]
		}
		case 14:{
			n := len(bbp.freeChunks) - 1
			r := bbp.freeChunks[n].(*[1024*14]byte)
			bbp.freeChunks[n] = nil
			bbp.freeChunks = bbp.freeChunks[:n]
			bbp.freeChunksLock.Unlock()
			return r[:]
		}
		case 15:{
			n := len(bbp.freeChunks) - 1
			r := bbp.freeChunks[n].(*[1024*15]byte)
			bbp.freeChunks[n] = nil
			bbp.freeChunks = bbp.freeChunks[:n]
			bbp.freeChunksLock.Unlock()
			return r[:]
		}
		default:
		fallthrough
		case 16:{
			n := len(bbp.freeChunks) - 1
			r := bbp.freeChunks[n].(*[1024*16]byte)
			bbp.freeChunks[n] = nil
			bbp.freeChunks = bbp.freeChunks[:n]
			bbp.freeChunksLock.Unlock()
			return r[:]
		}
	}
}

func (bbp *ByteBufferPool)Put(chunk []byte){
	if chunk == nil {
		return
	}
	chunk = chunk[:bbp.chunkSize]
	s := (bbp.chunkSize+1023)/1024
	switch(s){
		case 1:{
			p := (*[1024*1]byte)(unsafe.Pointer(&chunk[0]))
			bbp.freeChunksLock.Lock()
			bbp.freeChunks = append(bbp.freeChunks, p)
			bbp.freeChunksLock.Unlock()
		}
		case 2:{
			p := (*[1024*2]byte)(unsafe.Pointer(&chunk[0]))
			bbp.freeChunksLock.Lock()
			bbp.freeChunks = append(bbp.freeChunks, p)
			bbp.freeChunksLock.Unlock()
		}
		case 3:{
			p := (*[1024*3]byte)(unsafe.Pointer(&chunk[0]))
			bbp.freeChunksLock.Lock()
			bbp.freeChunks = append(bbp.freeChunks, p)
			bbp.freeChunksLock.Unlock()
		}
		case 4:{
			p := (*[1024*4]byte)(unsafe.Pointer(&chunk[0]))
			bbp.freeChunksLock.Lock()
			bbp.freeChunks = append(bbp.freeChunks, p)
			bbp.freeChunksLock.Unlock()
		}
		case 5:{
			p := (*[1024*5]byte)(unsafe.Pointer(&chunk[0]))
			bbp.freeChunksLock.Lock()
			bbp.freeChunks = append(bbp.freeChunks, p)
			bbp.freeChunksLock.Unlock()
		}
		case 6:{
			p := (*[1024*6]byte)(unsafe.Pointer(&chunk[0]))
			bbp.freeChunksLock.Lock()
			bbp.freeChunks = append(bbp.freeChunks, p)
			bbp.freeChunksLock.Unlock()
		}
		case 7:{
			p := (*[1024*7]byte)(unsafe.Pointer(&chunk[0]))
			bbp.freeChunksLock.Lock()
			bbp.freeChunks = append(bbp.freeChunks, p)
			bbp.freeChunksLock.Unlock()
		}
		case 8:{
			p := (*[1024*8]byte)(unsafe.Pointer(&chunk[0]))
			bbp.freeChunksLock.Lock()
			bbp.freeChunks = append(bbp.freeChunks, p)
			bbp.freeChunksLock.Unlock()
		}
		case 9:{
			p := (*[1024*9]byte)(unsafe.Pointer(&chunk[0]))
			bbp.freeChunksLock.Lock()
			bbp.freeChunks = append(bbp.freeChunks, p)
			bbp.freeChunksLock.Unlock()
		}
		case 10:{
			p := (*[1024*10]byte)(unsafe.Pointer(&chunk[0]))
			bbp.freeChunksLock.Lock()
			bbp.freeChunks = append(bbp.freeChunks, p)
			bbp.freeChunksLock.Unlock()
		}
		case 11:{
			p := (*[1024*11]byte)(unsafe.Pointer(&chunk[0]))
			bbp.freeChunksLock.Lock()
			bbp.freeChunks = append(bbp.freeChunks, p)
			bbp.freeChunksLock.Unlock()
		}
		case 12:{
			p := (*[1024*12]byte)(unsafe.Pointer(&chunk[0]))
			bbp.freeChunksLock.Lock()
			bbp.freeChunks = append(bbp.freeChunks, p)
			bbp.freeChunksLock.Unlock()
		}
		case 13:{
			p := (*[1024*13]byte)(unsafe.Pointer(&chunk[0]))
			bbp.freeChunksLock.Lock()
			bbp.freeChunks = append(bbp.freeChunks, p)
			bbp.freeChunksLock.Unlock()
		}
		case 14:{
			p := (*[1024*14]byte)(unsafe.Pointer(&chunk[0]))
			bbp.freeChunksLock.Lock()
			bbp.freeChunks = append(bbp.freeChunks, p)
			bbp.freeChunksLock.Unlock()
		}
		case 15:{
			p := (*[1024*15]byte)(unsafe.Pointer(&chunk[0]))
			bbp.freeChunksLock.Lock()
			bbp.freeChunks = append(bbp.freeChunks, p)
			bbp.freeChunksLock.Unlock()
		}
		default:
		fallthrough
		case 16:{
			p := (*[1024*16]byte)(unsafe.Pointer(&chunk[0]))
			bbp.freeChunksLock.Lock()
			bbp.freeChunks = append(bbp.freeChunks, p)
			bbp.freeChunksLock.Unlock()
		}
	}
}