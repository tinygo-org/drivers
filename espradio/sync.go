package espradio

// Various functions related to locks, mutexes, semaphores, and queues.

/*
#include "include.h"
*/
import "C"

import (
	"sync"
	"sync/atomic"
	"time"
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

type semaphore chan struct{}

var semaphores [2]semaphore
var semaphoreIndex uint32
var wifiSemaphore semaphore

//export espradio_semphr_create
func espradio_semphr_create(max, init uint32) unsafe.Pointer {
	newIndex := atomic.AddUint32(&semaphoreIndex, 1)
	sem := &semaphores[newIndex-1]
	ch := make(semaphore, max)
	for i := uint32(0); i < init; i++ {
		ch <- struct{}{}
	}
	*sem = ch
	return unsafe.Pointer(sem)
}

//export espradio_semphr_take
func espradio_semphr_take(semphr unsafe.Pointer, block_time_tick uint32) int32 {
	sem := (*semaphore)(semphr)
	if block_time_tick != C.OSI_FUNCS_TIME_BLOCKING {
		panic("espradio: todo: semphr_take with timeout")
	}
	<-*sem
	return 1
}

//export espradio_semphr_give
func espradio_semphr_give(semphr unsafe.Pointer) int32 {
	// Note: we might need to return 0 when sending isn't possible (e.g. using a
	// non-blocking send). According to the documentation of xSemaphoreGive:
	//
	// > pdTRUE if the semaphore was released. pdFALSE if an error occurred.
	// > Semaphores are implemented using queues. An error can occur if there is
	// > no space on the queue to post a message - indicating that the semaphore
	// > was not first obtained correctly.
	sem := (*semaphore)(semphr)
	*sem <- struct{}{}
	return 1
}

//export espradio_semphr_delete
func espradio_semphr_delete(semphr unsafe.Pointer) {
	sem := (*semaphore)(semphr)
	close(*sem)
}

//export espradio_wifi_thread_semphr_get
func espradio_wifi_thread_semphr_get() unsafe.Pointer {
	if wifiSemaphore == nil {
		wifiSemaphore = make(semaphore, 1)
	}
	return unsafe.Pointer(&wifiSemaphore)
}

type queueElementType [8]byte

//export espradio_wifi_create_queue
func espradio_wifi_create_queue(queue_len, item_size int) chan queueElementType {
	if item_size != len(queueElementType{}) {
		panic("espradio: unexpected queue item_size")
	}
	return make(chan queueElementType, item_size)
}

//export espradio_wifi_delete_queue
func espradio_wifi_delete_queue(queue chan queueElementType) {
	// We can't really delete a channel, but we can close it.
	close(queue)
}

//export espradio_queue_recv
func espradio_queue_recv(queue chan queueElementType, item unsafe.Pointer, block_time_tick uint32) int32 {
	// This is xQueueReceive.
	if block_time_tick != C.OSI_FUNCS_TIME_BLOCKING {
		panic("espradio: todo: queue_recv with timeout")
	}
	*(*[8]byte)(item) = <-queue
	return 1
}

//export espradio_queue_send
func espradio_queue_send(queue chan queueElementType, item unsafe.Pointer, block_time_tick uint32) int32 {
	// This is xQueueSend.
	if block_time_tick != C.OSI_FUNCS_TIME_BLOCKING {
		duration := time.Duration(ticksToMilliseconds(block_time_tick)) * time.Millisecond
		// TODO: reuse the timer to avoid allocating a new timer on each queue send
		select {
		case <-time.After(duration):
			return 0
		case queue <- *(*[8]byte)(item):
			return 1
		}
	}
	queue <- *(*[8]byte)(item)
	return 1
}
