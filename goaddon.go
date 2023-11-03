package main

import "C"
import "time"

//export HelloSync
func HelloSync(_name *C.char) *C.char {
	// 传入 string 类型，返回 string 类型
	name := C.GoString(_name)

	res := "hello"
	ch := make(chan bool)

	// 当使用协程时，由于 JS 使用同步式调用，JS 进程会发生阻塞等待返回
	go func() {
		// 耗时任务处理
		time.Sleep(time.Duration(2) * time.Second)
		if len(name) > 0 {
			res += "," + name
		}
		ch <- true
	}()

	<-ch

	return C.CString(res)
}

//export HelloAsync
func HelloAsync(_name *C.char, cbsFnName *C.char) *C.char {
	// 传入 string 类型，返回 string 类型
	name := C.GoString(_name)

	res := "hello"
	ch := make(chan bool)

	// 当使用协程时，由于 JS 使用异步式调用，JS 进程不会发生阻塞，当返回值时会 JS callback
	go func() {
		// 耗时任务处理
		time.Sleep(time.Duration(2) * time.Second)
		if len(name) > 0 {
			res += "," + name
		}
		ch <- true
	}()

	<-ch

	return C.CString(res)
}
