#include <stdint.h>
#include <stddef.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "espnet.h"
#include "esp_wifi.h"
#include "esp_private/wifi.h"
#include "freertos/FreeRTOS.h"
#include "freertos/semphr.h"
#include "freertos/task.h"


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

// Having conflict between when include 
// #include "freertos/portmacro.h" 

typedef struct {
	/* owner field values:
	 * 0				- Uninitialized (invalid)
	 * portMUX_FREE_VAL - Mux is free, can be locked by either CPU
	 * CORE_ID_REGVAL_PRO / CORE_ID_REGVAL_APP - Mux is locked to the particular core
	 *
	 *
	 * Any value other than portMUX_FREE_VAL, CORE_ID_REGVAL_PRO, CORE_ID_REGVAL_APP indicates corruption
	 */
	uint32_t owner;
	/* count field:
	 * If mux is unlocked, count should be zero.
	 * If mux is locked, count is non-zero & represents the number of recursive locks on the mux.
	 */
	uint32_t count;
} portMUX_TYPE;

#define portMUX_FREE_VAL	SPINLOCK_FREE
#define SPINLOCK_FREE	   0xB33FFFFF

#define portMUX_INITIALIZER_UNLOCKED {					\
		.owner = portMUX_FREE_VAL,						\
		.count = 0,										\
	}

static void * _spin_lock_create(void) {
	portMUX_TYPE tmp = portMUX_INITIALIZER_UNLOCKED;
	void *mux = malloc(sizeof(portMUX_TYPE));
	if (mux) {
		memcpy(mux,&tmp,sizeof(portMUX_TYPE));
		return mux;
	}
	return NULL;
}
static void _spin_lock_delete(void *lock) {
	free(lock);
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
	return (void *)xSemaphoreCreateCounting(max, init);
}
static void _semphr_delete(void *semphr) {
	vSemaphoreDelete(semphr);
}
static int32_t _semphr_take(void *semphr, uint32_t block_time_tick) {
	if (block_time_tick == OSI_FUNCS_TIME_BLOCKING) {
		return (int32_t)xSemaphoreTake(semphr, portMAX_DELAY);
	} else {
		return (int32_t)xSemaphoreTake(semphr, block_time_tick);
	}
}
static int32_t _semphr_give(void *semphr) {
	return (int32_t)xSemaphoreGive(semphr);
}
static void *_wifi_thread_semphr_get(void) {
	static SemaphoreHandle_t sem = NULL;
	if (!sem) {
		sem = xSemaphoreCreateCounting(1, 0);
	}
	return (void*)sem;
}
static void *_mutex_create(void) {
	printf("called: _mutex_create\n");
	return NULL;
}

static void *_recursive_mutex_create(void) {
	return xSemaphoreCreateRecursiveMutex();
}
static void _mutex_delete(void *mutex) {
	return vSemaphoreDelete(mutex);
}
static int32_t _mutex_lock(void *mutex) {
	return (int32_t)xSemaphoreTakeRecursive(mutex, portMAX_DELAY);
}
static int32_t _mutex_unlock(void *mutex) {
	return (int32_t)xSemaphoreGiveRecursive(mutex);
}
static void * _queue_create(uint32_t queue_len, uint32_t item_size) {
	printf("called: _queue_create\n");
	return NULL;
}
static void _queue_delete(void *queue) {
	printf("called: _queue_delete\n");
}
static int32_t _queue_send(void *queue, void *item, uint32_t block_time_tick) {
	if (block_time_tick == OSI_FUNCS_TIME_BLOCKING) {
		return (int32_t)xQueueSend(queue, item, portMAX_DELAY);
	} else {
		return (int32_t)xQueueSend(queue, item, block_time_tick);
	}
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
	if (block_time_tick == OSI_FUNCS_TIME_BLOCKING) {
		return (int32_t)xQueueReceive(queue, item, portMAX_DELAY);
	} else {
		return (int32_t)xQueueReceive(queue, item, block_time_tick);
	}
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

#define P(x) printf("called: "#x"\n");

static int32_t _task_create_pinned_to_core(void *task_func, const char *name, uint32_t stack_depth, void *param, uint32_t prio, void *task_handle, uint32_t core_id) {
	// Note: using xTaskCreate instead of xTaskCreatePinnedToCore.
	return (uint32_t)xTaskCreate(task_func, name, stack_depth, param, prio, task_handle);
}
static int32_t _task_create(void *task_func, const char *name, uint32_t stack_depth, void *param, uint32_t prio, void *task_handle) {
	P(_task_create)
	return 0;
}
static void _task_delete(void *task_handle) {
	P(_task_delete)
}
static int32_t _task_ms_to_tick(uint32_t ms) {
	return (int32_t)(ms / portTICK_PERIOD_MS);
}
static int32_t _task_get_max_priority() {
	return configMAX_PRIORITIES;
}
static int32_t _event_post(const char* event_base, int32_t event_id, void* event_data, size_t event_data_size, uint32_t ticks_to_wait) {
	P(_event_post)
	return 0;
}
static uint32_t  _get_free_heap_size(void) {
	P(_get_free_heap_size)
	return 0;
}
static uint32_t  _rand(void) {
	P(_rand)
	return 0;
}
static void _dport_access_stall_other_cpu_start_wrap(void) {
	P(_dport_access_stall_other_cpu_start_wrap)
}
static void _dport_access_stall_other_cpu_end_wrap(void) {
	P(_dport_access_stall_other_cpu_end_wrap)
}
static void _wifi_apb80m_request(void) {
	P(_wifi_apb80m_request)
}
static void _wifi_apb80m_release(void) {
	P(_wifi_apb80m_release)
}
static void _phy_disable(void) {
	P(_phy_disable)
}
static void _phy_enable(void) {
	P(_phy_enable)
}
static int _phy_update_country_info(const char* country) {
	P(_phy_update_country_info)
	return 0;
}
static int _read_mac(uint8_t* mac, uint32_t type) {
	P(_read_mac)
	return 0;
}
static void _timer_arm(void *timer, uint32_t tmout, bool repeat) {
	P(_timer_arm)
}
static void _timer_disarm(void *timer) {
	P(_timer_disarm)
}
static void _timer_done(void *ptimer) {
	P(_timer_done)
}
static void _timer_setfn(void *ptimer, void *pfunction, void *parg) {
	P(_timer_setfn)
}
static void _timer_arm_us(void *ptimer, uint32_t us, bool repeat) {
	P(_timer_arm_us)
}
static void _wifi_reset_mac(void) {
	P(_wifi_reset_mac)
}
static void _wifi_clock_enable(void) {
	P(_wifi_clock_enable)
}
static void _wifi_clock_disable(void) {
	P(_wifi_clock_disable)
}
static void _wifi_rtc_enable_iso(void) {
	P(_wifi_rtc_enable_iso)
}
static void _wifi_rtc_disable_iso(void) {
	P(_wifi_rtc_disable_iso)
}
static int64_t _esp_timer_get_time(void) {
	P(_esp_timer_get_time)
	return 0;
}
static int _nvs_set_i8(uint32_t handle, const char* key, int8_t value) {
	P(_nvs_set_i8)
	return 0;
}
static int _nvs_get_i8(uint32_t handle, const char* key, int8_t* out_value) {
	P(_nvs_get_i8)
	return 0;
}
static int _nvs_set_u8(uint32_t handle, const char* key, uint8_t value) {
	P(_nvs_set_u8)
	return 0;
}
static int _nvs_get_u8(uint32_t handle, const char* key, uint8_t* out_value) {
	P(_nvs_get_u8)
	return 0;
}
static int _nvs_set_u16(uint32_t handle, const char* key, uint16_t value) {
	P(_nvs_set_u16)
	return 0;
}
static int _nvs_get_u16(uint32_t handle, const char* key, uint16_t* out_value) {
	P(_nvs_get_u16)
	return 0;
}
static int _nvs_open(const char* name, uint32_t open_mode, uint32_t *out_handle) {
	P(_nvs_open)
	return 0;
}
static void _nvs_close(uint32_t handle) {
	P(_nvs_close)
}
static int _nvs_commit(uint32_t handle) {
	P(_nvs_commit)
	return 0;
}
static int _nvs_set_blob(uint32_t handle, const char* key, const void* value, size_t length) {
	P(_nvs_set_blob)
	return 0;
}
static int _nvs_get_blob(uint32_t handle, const char* key, void* out_value, size_t* length) {
	P(_nvs_get_blob)
	return 0;
}
static int _nvs_erase_key(uint32_t handle, const char* key) {
	P(_nvs_erase_key)
	return 0;
}
static int _get_random(uint8_t *buf, size_t len) {
	P(_get_random)
	return 0;
}
static int _get_time(void *t) {
	P(_get_time)
	return 0;
}
static unsigned long _random(void) {
	P(_random)
	return 0;
}
// #if CONFIG_IDF_TARGET_ESP32S2 || CONFIG_IDF_TARGET_ESP32S3 || CONFIG_IDF_TARGET_ESP32C3
//	 uint32_t (* _slowclk_cal_get(void)
// #endif
static void _log_write(uint32_t level, const char* tag, const char* format, ...) {
	va_list argList;
	printf("[%s] ", tag);
	va_start(argList, format);
	vprintf(format, argList);
	va_end(argList);
	printf("\n");
}
static void _log_writev(uint32_t level, const char* tag, const char* format, va_list args) {
	printf("[%s] ", tag);
	vprintf(format, args);
	printf("\n");
}

static uint32_t  _log_timestamp(void) {
	P(_log_timestamp)
	return 0;
}
static void* _malloc_internal(size_t size) {
	printf("called: _malloc_internal(%d)\n", size);
	return malloc(size);
}
static void* _realloc_internal(void *ptr, size_t size) {
	printf("called: _realloc_internal(%p,%d)\n", ptr, size);
	return NULL;
}

static void* _calloc_internal(size_t n, size_t size) {
	printf("called: _calloc_internal(%d,%d)\n", n, size);
	return malloc(n * size);
}
static void* _zalloc_internal(size_t size) {
	printf("called: _zalloc_internal(%d)\n", size);
	return NULL;
}
static void* _wifi_malloc(size_t size) {
	return malloc(size);
}
static void* _wifi_realloc(void *ptr, size_t size) {
	printf("called: _wifi_realloc(%d)\n", size);
	return NULL;
}
static void* _wifi_calloc(size_t n, size_t size) {
	return calloc(n, size);
}
static void* _wifi_zalloc(size_t size) {
	return calloc(1, size);
}
static void* _wifi_create_queue(int queue_len, int item_size) {
	wifi_static_queue_t *queue = (wifi_static_queue_t*)malloc(sizeof(wifi_static_queue_t));
	queue->handle = xQueueCreate( queue_len, item_size);
	return queue;
}
static void _wifi_delete_queue(void * queue) {
	vQueueDelete(queue);
}
static int _coex_init(void) {
	P(_coex_init)
	return 0;
}
static void _coex_deinit(void) {
	P(_coex_deinit)
}
static int _coex_enable(void) {
	P(_coex_enable)
	return 0;
}
static void _coex_disable(void) {
	P(_coex_disable)
}
static uint32_t  _coex_status_get(void) {
	P(_coex_status_get)
	return 0;
}
static void _coex_condition_set(uint32_t type, bool dissatisfy) {
	P(_coex_condition_set)
}
static int _coex_wifi_request(uint32_t event, uint32_t latency, uint32_t duration) {
	P(_coex_wifi_request)
	return 0;
}
static int _coex_wifi_release(uint32_t event) {
	P(_coex_wifi_release)
	return 0;
}
static int _coex_wifi_channel_set(uint8_t primary, uint8_t secondary) {
	P(_coex_wifi_channel_set)
	return 0;
}
static int _coex_event_duration_get(uint32_t event, uint32_t *duration) {
	P(_coex_event_duration_get)
	return 0;
}
static int _coex_pti_get(uint32_t event, uint8_t *pti) {
	P(_coex_pti_get)
	return 0;
}
static void _coex_schm_status_bit_clear(uint32_t type, uint32_t status) {
	P(_coex_schm_status_bit_clear)
}
static void _coex_schm_status_bit_set(uint32_t type, uint32_t status) {
	P(_coex_schm_status_bit_set)
}
static int _coex_schm_interval_set(uint32_t interval) {
	P(_coex_schm_interval_set)
	return 0;
}
static uint32_t  _coex_schm_interval_get(void) {
	P(_coex_schm_interval_get)
	return 0;
}
static uint8_t _coex_schm_curr_period_get(void) {
	P(_coex_schm_curr_period_get)
	return 0;
}
static void* _coex_schm_curr_phase_get(void) {
	P(_coex_schm_curr_phase_get)
	return NULL;
}
static int _coex_schm_curr_phase_idx_set(int idx) {
	P(_coex_schm_curr_phase_idx_set)
	return 0;
}
static int _coex_schm_curr_phase_idx_get(void) {
	P(_coex_schm_curr_phase_idx_get)
	return 0;
}


uint32_t _slowclk_cal_get(void) {
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
	._queue_msg_waiting = (uint32_t(*)(void *))uxQueueMessagesWaiting,
	._event_group_create = _event_group_create,
	._event_group_delete = _event_group_delete,
	._event_group_set_bits = _event_group_set_bits,
	._event_group_clear_bits = _event_group_clear_bits,
	._event_group_wait_bits = _event_group_wait_bits,
	._task_create_pinned_to_core = _task_create_pinned_to_core,
	._task_create = _task_create,
	._task_delete = _task_delete,
	._task_delay = vTaskDelay,
	._task_ms_to_tick = _task_ms_to_tick,
	._task_get_current_task = (void *(*)(void))xTaskGetCurrentTaskHandle,
	._task_get_max_priority = _task_get_max_priority,
	._malloc = malloc,
	._free = free,
	._event_post = _event_post,
	._get_free_heap_size = _get_free_heap_size,
	._rand = _rand,
	._dport_access_stall_other_cpu_start_wrap = _dport_access_stall_other_cpu_start_wrap,
	._dport_access_stall_other_cpu_end_wrap = _dport_access_stall_other_cpu_end_wrap,
	._wifi_apb80m_request = _wifi_apb80m_request,
	._wifi_apb80m_release = _wifi_apb80m_release,
	._phy_disable = _phy_disable,
	._phy_enable = _phy_enable,
	._phy_update_country_info = _phy_update_country_info,
	._read_mac = _read_mac,
	._timer_arm = _timer_arm,
	._timer_disarm = _timer_disarm,
	._timer_done = _timer_done,
	._timer_setfn = _timer_setfn,
	._timer_arm_us = _timer_arm_us,
	._wifi_reset_mac = _wifi_reset_mac,
	._wifi_clock_enable = _wifi_clock_enable,
	._wifi_clock_disable = _wifi_clock_disable,
	._wifi_rtc_enable_iso = _wifi_rtc_enable_iso,
	._wifi_rtc_disable_iso = _wifi_rtc_disable_iso,
	._esp_timer_get_time = _esp_timer_get_time,
	._nvs_set_i8 = _nvs_set_i8,
	._nvs_get_i8 = _nvs_get_i8,
	._nvs_set_u8 = _nvs_set_u8,
	._nvs_get_u8 = _nvs_get_u8,
	._nvs_set_u16 = _nvs_set_u16,
	._nvs_get_u16 = _nvs_get_u16,
	._nvs_open = _nvs_open,
	._nvs_close = _nvs_close,
	._nvs_commit = _nvs_commit,
	._nvs_set_blob = _nvs_set_blob,
	._nvs_get_blob = _nvs_get_blob,
	._nvs_erase_key = _nvs_erase_key,
	._get_random = _get_random,
	._get_time = _get_time,
	._random = _random,
#if CONFIG_IDF_TARGET_ESP32S2 || CONFIG_IDF_TARGET_ESP32S3 || CONFIG_IDF_TARGET_ESP32C3
	._slowclk_cal_get = _slowclk_cal_get,
#endif
	._log_write = _log_write,
	._log_writev = _log_writev,
	._log_timestamp = _log_timestamp,
	._malloc_internal = _malloc_internal,
	._realloc_internal = _realloc_internal,
	._calloc_internal = _calloc_internal,
	._zalloc_internal = _zalloc_internal,
	._wifi_malloc = _wifi_malloc,
	._wifi_realloc = _wifi_realloc,
	._wifi_calloc = _wifi_calloc,
	._wifi_zalloc = _wifi_zalloc,
	._wifi_create_queue = _wifi_create_queue,
	._wifi_delete_queue = _wifi_delete_queue,
	._coex_init = _coex_init,
	._coex_deinit = _coex_deinit,
	._coex_enable = _coex_enable,
	._coex_disable = _coex_disable,
	._coex_status_get = _coex_status_get,
	._coex_condition_set = _coex_condition_set,
	._coex_wifi_request = _coex_wifi_request,
	._coex_wifi_release = _coex_wifi_release,
	._coex_wifi_channel_set = _coex_wifi_channel_set,
	._coex_event_duration_get = _coex_event_duration_get,
	._coex_pti_get = _coex_pti_get,
	._coex_schm_status_bit_clear = _coex_schm_status_bit_clear,
	._coex_schm_status_bit_set = _coex_schm_status_bit_set,
	._coex_schm_interval_set = _coex_schm_interval_set,
	._coex_schm_interval_get = _coex_schm_interval_get,
	._coex_schm_curr_period_get = _coex_schm_curr_period_get,
	._coex_schm_curr_phase_get = _coex_schm_curr_phase_get,
	._coex_schm_curr_phase_idx_set = _coex_schm_curr_phase_idx_set,
	._coex_schm_curr_phase_idx_get = _coex_schm_curr_phase_idx_get,

	._magic = ESP_WIFI_OS_ADAPTER_MAGIC,
};

static int esp_aes_wrap(const unsigned char *kek, int n, const unsigned char *plain, unsigned char *cipher) {
	P(aes_wrap)
	return -1;
}
static int esp_aes_unwrap(const unsigned char *kek, int n, const unsigned char *cipher, unsigned char *plain) {
	P(aes_unwrap)
	return -1;	
}
static int hmac_sha256_vector(const unsigned char *key, int key_len, int num_elem,
	const unsigned char *addr[], const int *len, unsigned char *mac) {
	return -1;
}
static int sha256_prf(const unsigned char *key, int key_len, const char *label,
								   const unsigned char *data, int data_len, unsigned char *buf, int buf_len) {
	P(sha256_prf)
	return -1;
}
static int hmac_md5(const unsigned char *key, unsigned int key_len, const unsigned char *data,
							  unsigned int data_len, unsigned char *mac) {
	P(hmac_md5)
	return -1;
}
static int hamc_md5_vector(const unsigned char *key, unsigned int key_len, unsigned int num_elem,
							  const unsigned char *addr[], const unsigned int *len, unsigned char *mac) {
	P(hamc_md5_vector)
	return -1;							
}
static int hmac_sha1(const unsigned char *key, unsigned int key_len, const unsigned char *data,
							  unsigned int data_len, unsigned char *mac) {
	P(hmac_sha1)
	return -1;
}
static int hmac_sha1_vector(const unsigned char *key, unsigned int key_len, unsigned int num_elem,
							  const unsigned char *addr[], const unsigned int *len, unsigned char *mac) {
	P(hmac_sha1_vector)
	return -1;
}
static int sha1_prf(const unsigned char *key, unsigned int key_len, const char *label,
							  const unsigned char *data, unsigned int data_len, unsigned char *buf, unsigned int buf_len) {
	P(sha1_prf)
	return -1;
}
static int sha1_vector(unsigned int num_elem, const unsigned char *addr[], const unsigned int *len,
							  unsigned char *mac) {
	P(sha1_vector)
	return -1;
}
static int pbkdf2_sha1(const char *passphrase, const char *ssid, unsigned int ssid_len,
							  int iterations, unsigned char *buf, unsigned int buflen) {
	P(pbkdf2_sha1)
	return -1;
}
static int rc4_skip(const unsigned char *key, unsigned int keylen, unsigned int skip,
							  unsigned char *data, unsigned int data_len) {
	P(rc4_skip)
	return -1;
}
static int md5_vector(unsigned int num_elem, const unsigned char *addr[], const unsigned int *len,
							  unsigned char *mac) {
	P(md5_vector)
	return -1;
}
static void aes_encrypt(void *ctx, const unsigned char *plain, unsigned char *crypt) {
	P(aes_encrypt)
}
static void * aes_encrypt_init(const unsigned char *key,  unsigned int len) {
	P(aes_encrypt_init)
	return NULL;
}
static void aes_encrypt_deinit(void *ctx) {
	P(aes_encrypt_deinit)
}
static void aes_decrypt(void *ctx, const unsigned char *crypt, unsigned char *plain) {
	P(aes_decrypt)
}
static void * aes_decrypt_init(const unsigned char *key, unsigned int len) {
	P(aes_decrypt_init)
	return NULL;
}
static void aes_decrypt_deinit(void *ctx) {
	P(aes_decrypt_deinit)
}
static int aes_128_decrypt(const unsigned char *key, const unsigned char *iv, unsigned char *data, int data_len) {
	P(aes_128_decrypt)
	return -1;
}
static int omac1_aes_128(const uint8_t *key, const uint8_t *data, size_t data_len,
								   uint8_t *mic) {
	P(omac1_aes_128)
	return -1;
}
static uint8_t * ccmp_decrypt(const uint8_t *tk, const uint8_t *ieee80211_hdr,
										const uint8_t *data, size_t data_len,
										size_t *decrypted_len, bool espnow_pkt) {
	P(ccmp_decrypt)
	return NULL;
}
static uint8_t * ccmp_encrypt(const uint8_t *tk, uint8_t *frame, size_t len, size_t hdrlen,
										uint8_t *pn, int keyid, size_t *encrypted_len) {
	P(ccmp_encrypt)
	return NULL;
}
static int hmac_md5_vector(const unsigned char *key, unsigned int key_len, unsigned int num_elem,
							  const unsigned char *addr[], const unsigned int *len, unsigned char *mac) {
	P(hmac_md5_vector)
	return -1;
}
static void esp_aes_encrypt(void *ctx, const unsigned char *plain, unsigned char *crypt) {
	P(esp_aes_encrypt)
}
static void esp_aes_decrypt(void *ctx, const unsigned char *crypt, unsigned char *plain) {
	P(esp_aes_decrypt)
}
static int aes_128_cbc_encrypt(const unsigned char *key, const unsigned char *iv, unsigned char *data, int data_len) {
	P(aes_128_cbc_encrypt)
	return -1;
}
static int aes_128_cbc_decrypt(const unsigned char *key, const unsigned char *iv, unsigned char *data, int data_len) {
	P(aes_128_cbc_decrypt)
	return -1;
}

const wpa_crypto_funcs_t g_wifi_default_wpa_crypto_funcs = {
	.size = sizeof(wpa_crypto_funcs_t),
	.version = ESP_WIFI_CRYPTO_VERSION,
	.aes_wrap = (esp_aes_wrap_t)esp_aes_wrap,
	.aes_unwrap = (esp_aes_unwrap_t)esp_aes_unwrap,
	.hmac_sha256_vector = (esp_hmac_sha256_vector_t)hmac_sha256_vector,
	.sha256_prf = (esp_sha256_prf_t)sha256_prf,
	.hmac_md5 = (esp_hmac_md5_t)hmac_md5,
	.hamc_md5_vector = (esp_hmac_md5_vector_t)hmac_md5_vector,
	.hmac_sha1 = (esp_hmac_sha1_t)hmac_sha1,
	.hmac_sha1_vector = (esp_hmac_sha1_vector_t)hmac_sha1_vector,
	.sha1_prf = (esp_sha1_prf_t)sha1_prf,
	.sha1_vector = (esp_sha1_vector_t)sha1_vector,
	.pbkdf2_sha1 = (esp_pbkdf2_sha1_t)pbkdf2_sha1,
	.rc4_skip = (esp_rc4_skip_t)rc4_skip,
	.md5_vector = (esp_md5_vector_t)md5_vector,
	.aes_encrypt = (esp_aes_encrypt_t)esp_aes_encrypt,
	.aes_encrypt_init = (esp_aes_encrypt_init_t)aes_encrypt_init,
	.aes_encrypt_deinit = (esp_aes_encrypt_deinit_t)aes_encrypt_deinit,
	.aes_decrypt = (esp_aes_decrypt_t)esp_aes_decrypt,
	.aes_decrypt_init = (esp_aes_decrypt_init_t)aes_decrypt_init,
	.aes_decrypt_deinit = (esp_aes_decrypt_deinit_t)aes_decrypt_deinit,
	.aes_128_encrypt = (esp_aes_128_encrypt_t)aes_128_cbc_encrypt,
	.aes_128_decrypt = (esp_aes_128_decrypt_t)aes_128_cbc_decrypt,
	.omac1_aes_128 = (esp_omac1_aes_128_t)omac1_aes_128,
	.ccmp_decrypt = (esp_ccmp_decrypt_t)ccmp_decrypt,
	.ccmp_encrypt = (esp_ccmp_encrypt_t)ccmp_encrypt
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
