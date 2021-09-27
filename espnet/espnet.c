#include <stdint.h>
#include <stddef.h>
#include <stdio.h>
#include "espnet.h"

// Stub functions, to know which functions need to be implemented for OS
// functionality.

static bool _env_is_chip(void) {
    printf("called: _env_is_chip\n");
	return false;
}
static void _set_intr(int32_t cpu_no, uint32_t intr_source, uint32_t intr_num, int32_t intr_prio) {
    printf("called: _set_intr\n");
}
static void _clear_intr(uint32_t intr_source, uint32_t intr_num) {
    printf("called: _clear_intr\n");
}
static void _set_isr(int32_t n, void *f, void *arg) {
    printf("called: _set_isr\n");
}
static void _ints_on(uint32_t mask) {
    printf("called: _ints_on\n");
}
static void _ints_off(uint32_t mask) {
    printf("called: _ints_off\n");
}
static bool _is_from_isr(void) {
    printf("called: _is_from_isr\n");
	return false;
}
static void * _spin_lock_create(void) {
    printf("called: _spin_lock_create\n");
	return NULL;
}
static void _spin_lock_delete(void *lock) {
    printf("called: _spin_lock_delete\n");
}
static uint32_t _wifi_int_disable(void *wifi_int_mux) {
    printf("called: _wifi_int_disable\n");
	return 0;
}
static void _wifi_int_restore(void *wifi_int_mux, uint32_t tmp) {
    printf("called: _wifi_int_restore\n");
}
static void _task_yield_from_isr(void) {
    printf("called: _task_yield_from_isr\n");
}
static void *_semphr_create(uint32_t max, uint32_t init) {
    printf("called: _semphr_create\n");
	return NULL;
}
static void _semphr_delete(void *semphr) {
    printf("called: _semphr_delete\n");
}
static int32_t _semphr_take(void *semphr, uint32_t block_time_tick) {
    printf("called: _semphr_take\n");
	return 0;
}
static int32_t _semphr_give(void *semphr) {
    printf("called: _semphr_give\n");
	return 0;
}
static void *_wifi_thread_semphr_get(void) {
    printf("called: _wifi_thread_semphr_get\n");
	return NULL;
}
static void *_mutex_create(void) {
    printf("called: _mutex_create\n");
	return NULL;
}
static void *_recursive_mutex_create(void) {
    printf("called: _recursive_mutex_create\n");
	return NULL;
}
static void _mutex_delete(void *mutex) {
    printf("called: _mutex_delete\n");
}
static int32_t _mutex_lock(void *mutex) {
    printf("called: _mutex_lock\n");
	return 0;
}
static int32_t _mutex_unlock(void *mutex) {
    printf("called: _mutex_unlock\n");
	return 0;
}
static void * _queue_create(uint32_t queue_len, uint32_t item_size) {
    printf("called: _queue_create\n");
	return NULL;
}
static void _queue_delete(void *queue) {
    printf("called: _queue_delete\n");
}
static int32_t _queue_send(void *queue, void *item, uint32_t block_time_tick) {
    printf("called: _queue_send\n");
	return 0;
}
static int32_t _queue_send_from_isr(void *queue, void *item, void *hptw) {
    printf("called: _queue_send_from_isr\n");
	return 0;
}
static int32_t _queue_send_to_back(void *queue, void *item, uint32_t block_time_tick) {
    printf("called: _queue_send_to_back\n");
	return 0;
}
static int32_t _queue_send_to_front(void *queue, void *item, uint32_t block_time_tick) {
    printf("called: _queue_send_to_front\n");
	return 0;
}
static int32_t _queue_recv(void *queue, void *item, uint32_t block_time_tick) {
    printf("called: _queue_recv\n");
	return 0;
}
static uint32_t _queue_msg_waiting(void *queue) {
    printf("called: _queue_msg_waiting\n");
	return 0;
}
static void * _event_group_create(void) {
    printf("called: _event_group_create\n");
	return NULL;
}
static void _event_group_delete(void *event) {
    printf("called: _event_group_delete\n");
}
static uint32_t _event_group_set_bits(void *event, uint32_t bits) {
    printf("called: _event_group_set_bits\n");
	return 0;
}
static uint32_t _event_group_clear_bits(void *event, uint32_t bits) {
    printf("called: _event_group_clear_bits\n");
	return 0;
}
static uint32_t _event_group_wait_bits(void *event, uint32_t bits_to_wait_for, int clear_on_exit, int wait_for_all_bits, uint32_t block_time_tick) {
    printf("called: _event_group_wait_bits\n");
	return 0;
}

// OS adapter functions.
// See: esp-idf/components/esp_wifi/include/esp_private/wifi_os_adapter.h
wifi_osi_funcs_t g_wifi_osi_funcs = {
    ._version = ESP_WIFI_OS_ADAPTER_VERSION,
    ._env_is_chip = _env_is_chip,
    ._set_intr = _set_intr,
    ._clear_intr = _clear_intr,
    ._set_isr = _set_isr,
    ._ints_on = _ints_on,
    ._ints_off = _ints_off,
    ._is_from_isr = _is_from_isr,
    ._spin_lock_create = _spin_lock_create,
    ._spin_lock_delete = _spin_lock_delete,
    ._wifi_int_disable = _wifi_int_disable,
    ._wifi_int_restore = _wifi_int_restore,
    ._task_yield_from_isr = _task_yield_from_isr,
    ._semphr_create = _semphr_create,
    ._semphr_delete = _semphr_delete,
    ._semphr_take = _semphr_take,
    ._semphr_give = _semphr_give,
    ._wifi_thread_semphr_get = _wifi_thread_semphr_get,
    ._mutex_create = _mutex_create,
    ._recursive_mutex_create = _recursive_mutex_create,
    ._mutex_delete = _mutex_delete,
    ._mutex_lock = _mutex_lock,
    ._mutex_unlock = _mutex_unlock,
    ._queue_create = _queue_create,
    ._queue_delete = _queue_delete,
    ._queue_send = _queue_send,
    ._queue_send_from_isr = _queue_send_from_isr,
    ._queue_send_to_back = _queue_send_to_back,
    ._queue_send_to_front = _queue_send_to_front,
    ._queue_recv = _queue_recv,
    ._queue_msg_waiting = _queue_msg_waiting,
    ._event_group_create = _event_group_create,
    ._event_group_delete = _event_group_delete,
    ._event_group_set_bits = _event_group_set_bits,
    ._event_group_clear_bits = _event_group_clear_bits,
    ._event_group_wait_bits = _event_group_wait_bits,
    // TODO: define more of these functions
    ._magic = ESP_WIFI_OS_ADAPTER_MAGIC,
};

// This is a string constant that is used all over ESP-IDF and is also used by
// libnet80211.a. The main purpose is to be a fixed pointer that can be compared
// against etc.
const char *WIFI_EVENT = "WIFI_EVENT";

// Required by libphy.a
int phy_printf(const char *format, ...) {
    va_list args;
    va_start(args, format);
    printf("phy: ");
    int res = vprintf(format, args);
    va_end(args);
    return res;
}

// Required by libpp.a
int pp_printf(const char *format, ...) {
    va_list args;
    va_start(args, format);
    printf("pp: ");
    int res = vprintf(format, args);
    va_end(args);
    return res;
}

// Required by libnet80211.a
int net80211_printf(const char *format, ...) {
    va_list args;
    va_start(args, format);
    printf("net80211: ");
    int res = vprintf(format, args);
    va_end(args);
    return res;
}

// Source: esp-idf/components/wpa_supplicant/src/utils/common.c
static int hex2num(char c)
{
	if (c >= '0' && c <= '9')
		return c - '0';
	if (c >= 'a' && c <= 'f')
		return c - 'a' + 10;
	if (c >= 'A' && c <= 'F')
		return c - 'A' + 10;
	return -1;
}

// Source: esp-idf/components/wpa_supplicant/src/utils/common.c
int hex2byte(const char *hex)
{
	int a, b;
	a = hex2num(*hex++);
	if (a < 0)
		return -1;
	b = hex2num(*hex++);
	if (b < 0)
		return -1;
	return (a << 4) | b;
}

// Source: esp-idf/components/wpa_supplicant/src/utils/common.c
/**
 * hexstr2bin - Convert ASCII hex string into binary data
 * @hex: ASCII hex string (e.g., "01ab")
 * @buf: Buffer for the binary data
 * @len: Length of the text to convert in bytes (of buf); hex will be double
 * this size
 * Returns: 0 on success, -1 on failure (invalid hex string)
 */
int hexstr2bin(const char *hex, uint8_t *buf, size_t len)
{
	size_t i;
	int a;
	const char *ipos = hex;
	uint8_t *opos = buf;

	for (i = 0; i < len; i++) {
		a = hex2byte(ipos);
		if (a < 0)
			return -1;
		*opos++ = a;
		ipos += 2;
	}
	return 0;
}
