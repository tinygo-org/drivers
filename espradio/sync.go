package espradio

// Various functions related to locks, mutexes, semaphores, and queues.

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

// Use a single fake spinlock. This is also how the Rust port does it.
var fakeSpinLock uint8

//export espradio_spin_lock_create
func espradio_spin_lock_create() unsafe.Pointer {
	return unsafe.Pointer(&fakeSpinLock)
}

//export espradio_spin_lock_delete
func espradio_spin_lock_delete(lock unsafe.Pointer) {
}

// Use a global array of mutexes, because the binary blobs don't need that many.
var mutexes [2]sync.Mutex
var mutexIndex uint32

//export espradio_recursive_mutex_create
func espradio_recursive_mutex_create() unsafe.Pointer {
	// Allocate a mutex from the global array.
	// If the radio needs more mutexes, this will result in an index out of
	// range panic (hopefully including a source location, if using
	// `tinygo flash -monitor`).
	newIndex := atomic.AddUint32(&mutexIndex, 1)
	mutex := &mutexes[newIndex-1]
	return unsafe.Pointer(mutex)
}

//export espradio_mutex_lock
func espradio_mutex_lock(cmut unsafe.Pointer) int32 {
	// This is xSemaphoreTake with an infinite timeout in ESP-IDF. Therefore,
	// just lock the mutex and return true.
	// TODO: recursive locking. See:
	// https://www.freertos.org/RTOS-Recursive-Mutexes.html
	// For that we need to track the current goroutine - or maybe just whether
	// we're inside a special goroutine like the timer goroutine.
	mut := (*sync.Mutex)(cmut)
	mut.Lock()
	return 1
}

//export espradio_mutex_unlock
func espradio_mutex_unlock(cmut unsafe.Pointer) int32 {
	// Note: this is xSemaphoreGive in the ESP-IDF, which doesn't panic when
	// unlocking fails but rather returns false.
	mut := (*sync.Mutex)(cmut)
	mut.Unlock()
	return 1
}

var semaphores [1]uint32
var semaphoreIndex uint32

//export espradio_semphr_create
func espradio_semphr_create(max, init uint32) unsafe.Pointer {
	newIndex := atomic.AddUint32(&semaphoreIndex, 1)
	sem := &semaphores[newIndex-1]
	return unsafe.Pointer(sem)
}

type queueElementType [8]byte

// TODO: I think the return type results in undefined behavior (but it's still a
// pointer so I hope LLVM won't use that fact).
//
//export espradio_wifi_create_queue
func espradio_wifi_create_queue(queue_len, item_size int) chan queueElementType {
	if item_size > len(queueElementType{}) {
		panic("espradio: queue item_size too large")
	}
	return make(chan queueElementType, item_size)
}
