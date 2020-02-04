package utils
// Copyright 2018 chelion. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
import(
	"time"
	"unsafe"
	"runtime"
	"crypto/md5"
	"hash/crc32"
	"sync/atomic"
	"encoding/hex"
	"encoding/base64"
)

var(
	nowUnixSec int64
)

func init(){
	t := time.Now().Unix()
	atomic.StoreInt64(&nowUnixSec,t)
	go func() {
		for {
			time.Sleep(1*time.Second)
			atomic.AddInt64(&nowUnixSec,1)
		}
	}()
}

func GetNowUnixSec()int64{
	return atomic.LoadInt64(&nowUnixSec)
}

func IsWindows()(sta bool){
	if runtime.GOOS == "windows" {
		return true
	}
	return false
}

func Int32ToBytes(n int32) []byte {
	lenBytes := make([]byte, 4)
	lenBytes[0] = byte(n >> 24)
	lenBytes[1] = byte(n >> 16)
	lenBytes[2] = byte(n >> 8)
	lenBytes[3] = byte(n)
	return lenBytes
}

func BytesToInt32(n []byte) int32 {
	return int32(int32(n[0])<<24) + int32(int32(n[1])<<16) + int32(int32(n[2])<<8) + int32(n[3])
}

func GetStats()(*runtime.MemStats){
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)
	return memStats
}

func Base64URLEncode(in []byte)(out string,err error){
	out = base64.URLEncoding.EncodeToString([]byte(in))
	return out,nil
}

func Base64URLDecode(in string)(out []byte,err error){
	out,err = base64.URLEncoding.DecodeString(in)
	return
}

func String2Bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func Bytes2String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func MD5(str string) string {
    h := md5.New()
    h.Write([]byte(str))
    return hex.EncodeToString(h.Sum(nil))
}

func CRC32(v []byte)(uint32){
	return crc32.ChecksumIEEE(v)
}