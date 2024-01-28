//go:build esp32c3

#include "include.h"
#include <stdarg.h>

// Documentation for these functions:
// https://github.com/esp-rs/esp-wifi/blob/main/esp-wifi/src/wifi/os_adapter.rs

__attribute__((noreturn))
void espradio_panic(char *s);

static bool espradio_env_is_chip(void) {
    espradio_panic("todo: _env_is_chip");
}

static void espradio_set_intr(int32_t cpu_no, uint32_t intr_source, uint32_t intr_num, int32_t intr_prio) {
    espradio_panic("todo: _set_intr");
}

static void espradio_clear_intr(uint32_t intr_source, uint32_t intr_num) {
    espradio_panic("todo: _clear_intr");
}

static void espradio_set_isr(int32_t n, void *f, void *arg) {
    espradio_panic("todo: _set_isr");
}

static void espradio_ints_on(uint32_t mask) {
    espradio_panic("todo: _ints_on");
}

static void espradio_ints_off(uint32_t mask) {
    espradio_panic("todo: _ints_off");
}

static bool espradio_is_from_isr(void) {
    espradio_panic("todo: _is_from_isr");
}

void *espradio_spin_lock_create(void);

void espradio_spin_lock_delete(void *lock);

uint32_t espradio_wifi_int_disable(void *wifi_int_mux);

void espradio_wifi_int_restore(void *wifi_int_mux, uint32_t tmp);

static void espradio_task_yield_from_isr(void) {
    espradio_panic("todo: _task_yield_from_isr");
}

void *espradio_semphr_create(uint32_t max, uint32_t init);

void espradio_semphr_delete(void *semphr);

int32_t espradio_semphr_take(void *semphr, uint32_t block_time_tick);

int32_t espradio_semphr_give(void *semphr);

void *espradio_wifi_thread_semphr_get(void);

static void *espradio_mutex_create(void) {
    espradio_panic("todo: _mutex_create");
}

void *espradio_recursive_mutex_create(void);

static void espradio_mutex_delete(void *mutex) {
    espradio_panic("todo: _mutex_delete");
}

int32_t espradio_mutex_lock(void *mutex);

int32_t espradio_mutex_unlock(void *mutex);

static void *espradio_queue_create(uint32_t queue_len, uint32_t item_size) {
    espradio_panic("todo: _queue_create");
}

static void espradio_queue_delete(void *queue) {
    espradio_panic("todo: _queue_delete");
}

int32_t espradio_queue_send(void *queue, void *item, uint32_t block_time_tick);

static int32_t espradio_queue_send_from_isr(void *queue, void *item, void *hptw) {
    espradio_panic("todo: _queue_send_from_isr");
}

static int32_t espradio_queue_send_to_back(void *queue, void *item, uint32_t block_time_tick) {
    espradio_panic("todo: _queue_send_to_back");
}

static int32_t espradio_queue_send_to_front(void *queue, void *item, uint32_t block_time_tick) {
    espradio_panic("todo: _queue_send_to_front");
}

int32_t espradio_queue_recv(void *queue, void *item, uint32_t block_time_tick);

static uint32_t espradio_queue_msg_waiting(void *queue) {
    espradio_panic("todo: _queue_msg_waiting");
}

static void *espradio_event_group_create(void) {
    espradio_panic("todo: _event_group_create");
}

static void espradio_event_group_delete(void *event) {
    espradio_panic("todo: _event_group_delete");
}

static uint32_t espradio_event_group_set_bits(void *event, uint32_t bits) {
    espradio_panic("todo: _event_group_set_bits");
}

static uint32_t espradio_event_group_clear_bits(void *event, uint32_t bits) {
    espradio_panic("todo: _event_group_clear_bits");
}

static uint32_t espradio_event_group_wait_bits(void *event, uint32_t bits_to_wait_for, int clear_on_exit, int wait_for_all_bits, uint32_t block_time_tick) {
    espradio_panic("todo: _event_group_wait_bits");
}

void espradio_run_task(void *task_func, void *task_handle) {
    void (*fn)(void *task_handle) = task_func;
    fn(task_handle);
}

int32_t espradio_task_create_pinned_to_core(void *task_func, const char *name, uint32_t stack_depth, void *param, uint32_t prio, void *task_handle, uint32_t core_id);

static int32_t espradio_task_create(void *task_func, const char *name, uint32_t stack_depth, void *param, uint32_t prio, void *task_handle) {
    espradio_panic("todo: _task_create");
}

void espradio_task_delete(void *task_handle);

void espradio_task_delay(uint32_t tick);

int32_t espradio_task_ms_to_tick(uint32_t ms);

void *espradio_task_get_current_task(void);

static int32_t espradio_task_get_max_priority(void) {
    return 255; // arbitrary number
}

static void *espradio_malloc(size_t size) {
    espradio_panic("todo: _malloc");
}

static void espradio_free(void *p) {
    free(p);
}

static int32_t espradio_event_post(const char* event_base, int32_t event_id, void* event_data, size_t event_data_size, uint32_t ticks_to_wait) {
    espradio_panic("todo: _event_post");
}

static uint32_t espradio_get_free_heap_size(void) {
    espradio_panic("todo: _get_free_heap_size");
}

static uint32_t espradio_rand(void) {
    espradio_panic("todo: _rand");
}

static void espradio_dport_access_stall_other_cpu_start_wrap(void) {
    espradio_panic("todo: _dport_access_stall_other_cpu_start_wrap");
}

static void espradio_dport_access_stall_other_cpu_end_wrap(void) {
    espradio_panic("todo: _dport_access_stall_other_cpu_end_wrap");
}

static void espradio_wifi_apb80m_request(void) {
    espradio_panic("todo: _wifi_apb80m_request");
}

static void espradio_wifi_apb80m_release(void) {
    espradio_panic("todo: _wifi_apb80m_release");
}

static void espradio_phy_disable(void) {
    espradio_panic("todo: _phy_disable");
}

static void espradio_phy_enable(void) {
    espradio_panic("todo: _phy_enable");
}

static int espradio_phy_update_country_info(const char* country) {
    espradio_panic("todo: _phy_update_country_info");
}

static int espradio_read_mac(uint8_t* mac, unsigned int type) {
    espradio_panic("todo: _read_mac");
}

static void espradio_timer_arm(void *timer, uint32_t tmout, bool repeat) {
    espradio_panic("todo: _timer_arm");
}

static void espradio_timer_disarm(void *timer) {
    espradio_panic("todo: _timer_disarm");
}

static void espradio_timer_done(void *ptimer) {
    espradio_panic("todo: _timer_done");
}

static void espradio_timer_setfn(void *ptimer, void *pfunction, void *parg) {
    espradio_panic("todo: _timer_setfn");
}

static void espradio_timer_arm_us(void *ptimer, uint32_t us, bool repeat) {
    espradio_panic("todo: _timer_arm_us");
}

static void espradio_wifi_reset_mac(void) {
    espradio_panic("todo: _wifi_reset_mac");
}

static void espradio_wifi_clock_enable(void) {
    espradio_panic("todo: _wifi_clock_enable");
}

static void espradio_wifi_clock_disable(void) {
    espradio_panic("todo: _wifi_clock_disable");
}

static void espradio_wifi_rtc_enable_iso(void) {
    espradio_panic("todo: _wifi_rtc_enable_iso");
}

static void espradio_wifi_rtc_disable_iso(void) {
    espradio_panic("todo: _wifi_rtc_disable_iso");
}

static int64_t espradio_esp_timer_get_time(void) {
    espradio_panic("todo: _esp_timer_get_time");
}

static int espradio_nvs_set_i8(uint32_t handle, const char* key, int8_t value) {
    espradio_panic("todo: _nvs_set_i8");
}

static int espradio_nvs_get_i8(uint32_t handle, const char* key, int8_t* out_value) {
    espradio_panic("todo: _nvs_get_i8");
}

static int espradio_nvs_set_u8(uint32_t handle, const char* key, uint8_t value) {
    espradio_panic("todo: _nvs_set_u8");
}

static int espradio_nvs_get_u8(uint32_t handle, const char* key, uint8_t* out_value) {
    espradio_panic("todo: _nvs_get_u8");
}

static int espradio_nvs_set_u16(uint32_t handle, const char* key, uint16_t value) {
    espradio_panic("todo: _nvs_set_u16");
}

static int espradio_nvs_get_u16(uint32_t handle, const char* key, uint16_t* out_value) {
    espradio_panic("todo: _nvs_get_u16");
}

static int espradio_nvs_open(const char* name, unsigned int open_mode, uint32_t *out_handle) {
    espradio_panic("todo: _nvs_open");
}

static void espradio_nvs_close(uint32_t handle) {
    espradio_panic("todo: _nvs_close");
}

static int espradio_nvs_commit(uint32_t handle) {
    espradio_panic("todo: _nvs_commit");
}

static int espradio_nvs_set_blob(uint32_t handle, const char* key, const void* value, size_t length) {
    espradio_panic("todo: _nvs_set_blob");
}

static int espradio_nvs_get_blob(uint32_t handle, const char* key, void* out_value, size_t* length) {
    espradio_panic("todo: _nvs_get_blob");
}

static int espradio_nvs_erase_key(uint32_t handle, const char* key) {
    espradio_panic("todo: _nvs_erase_key");
}

static int espradio_get_random(uint8_t *buf, size_t len) {
    espradio_panic("todo: _get_random");
}

static int espradio_get_time(void *t) {
    espradio_panic("todo: _get_time");
}

static unsigned long espradio_random(void) {
    espradio_panic("todo: _random");
}

static uint32_t espradio_slowclk_cal_get(void) {
    espradio_panic("todo: _slowclk_cal_get");
}

static void espradio_log_writev(unsigned int level, const char* tag, const char* format, va_list args) {
    // Note: 'level' and 'tag' may be used to filter log messages.
    vprintf(format, args);
}

static void espradio_log_write(unsigned int level, const char* tag, const char* format, ...) {
    va_list args;
    va_start(args, format);
    espradio_log_writev(level, tag, format, args);
    va_end(args);
}

uint32_t espradio_log_timestamp(void);

static void * espradio_malloc_internal(size_t size) {
    return malloc(size);
}

static void * espradio_realloc_internal(void *ptr, size_t size) {
    espradio_panic("todo: _realloc_internal");
}

static void * espradio_calloc_internal(size_t n, size_t size) {
    return calloc(n, size);
}

static void * espradio_zalloc_internal(size_t size) {
    espradio_panic("todo: _zalloc_internal");
}

static void * espradio_wifi_malloc(size_t size) {
    espradio_panic("todo: _wifi_malloc");
}

static void * espradio_wifi_realloc(void *ptr, size_t size) {
    espradio_panic("todo: _wifi_realloc");
}

static void * espradio_wifi_calloc(size_t n, size_t size) {
    espradio_panic("todo: _wifi_calloc");
}

static void * espradio_wifi_zalloc(size_t size) {
    return calloc(1, size);
}

void * espradio_wifi_create_queue(int queue_len, int item_size);

void espradio_wifi_delete_queue(void * queue);

static int espradio_coex_init(void) {
    espradio_panic("todo: _coex_init");
}

static void espradio_coex_deinit(void) {
    espradio_panic("todo: _coex_deinit");
}

static int espradio_coex_enable(void) {
    espradio_panic("todo: _coex_enable");
}

static void espradio_coex_disable(void) {
    espradio_panic("todo: _coex_disable");
}

static uint32_t espradio_coex_status_get(void) {
    espradio_panic("todo: _coex_status_get");
}

static void espradio_coex_condition_set(uint32_t type, bool dissatisfy) {
    espradio_panic("todo: _coex_condition_set");
}

static int espradio_coex_wifi_request(uint32_t event, uint32_t latency, uint32_t duration) {
    espradio_panic("todo: _coex_wifi_request");
}

static int espradio_coex_wifi_release(uint32_t event) {
    espradio_panic("todo: _coex_wifi_release");
}

static int espradio_coex_wifi_channel_set(uint8_t primary, uint8_t secondary) {
    espradio_panic("todo: _coex_wifi_channel_set");
}

static int espradio_coex_event_duration_get(uint32_t event, uint32_t *duration) {
    espradio_panic("todo: _coex_event_duration_get");
}

static int espradio_coex_pti_get(uint32_t event, uint8_t *pti) {
    espradio_panic("todo: _coex_pti_get");
}

static void espradio_coex_schm_status_bit_clear(uint32_t type, uint32_t status) {
    espradio_panic("todo: _coex_schm_status_bit_clear");
}

static void espradio_coex_schm_status_bit_set(uint32_t type, uint32_t status) {
    espradio_panic("todo: _coex_schm_status_bit_set");
}

static int espradio_coex_schm_interval_set(uint32_t interval) {
    espradio_panic("todo: _coex_schm_interval_set");
}

static uint32_t espradio_coex_schm_interval_get(void) {
    espradio_panic("todo: _coex_schm_interval_get");
}

static uint8_t espradio_coex_schm_curr_period_get(void) {
    espradio_panic("todo: _coex_schm_curr_period_get");
}

static void * espradio_coex_schm_curr_phase_get(void) {
    espradio_panic("todo: _coex_schm_curr_phase_get");
}

static int espradio_coex_schm_process_restart(void) {
    espradio_panic("todo: _coex_schm_process_restart");
}

static int espradio_coex_schm_register_cb(int type, int (* cb)(int)) {
    espradio_panic("todo: _coex_schm_register_cb");
}

static int espradio_coex_register_start_cb(int (* cb)(void)) {
    espradio_panic("todo: _coex_register_start_cb");
}


wifi_osi_funcs_t espradio_osi_funcs = {
    ._version = ESP_WIFI_OS_ADAPTER_VERSION,
    ._env_is_chip = espradio_env_is_chip,
    ._set_intr = espradio_set_intr,
    ._clear_intr = espradio_clear_intr,
    ._set_isr = espradio_set_isr,
    ._ints_on = espradio_ints_on,
    ._ints_off = espradio_ints_off,
    ._is_from_isr = espradio_is_from_isr,
    ._spin_lock_create = espradio_spin_lock_create,
    ._spin_lock_delete = espradio_spin_lock_delete,
    ._wifi_int_disable = espradio_wifi_int_disable,
    ._wifi_int_restore = espradio_wifi_int_restore,
    ._task_yield_from_isr = espradio_task_yield_from_isr,
    ._semphr_create = espradio_semphr_create,
    ._semphr_delete = espradio_semphr_delete,
    ._semphr_take = espradio_semphr_take,
    ._semphr_give = espradio_semphr_give,
    ._wifi_thread_semphr_get = espradio_wifi_thread_semphr_get,
    ._mutex_create = espradio_mutex_create,
    ._recursive_mutex_create = espradio_recursive_mutex_create,
    ._mutex_delete = espradio_mutex_delete,
    ._mutex_lock = espradio_mutex_lock,
    ._mutex_unlock = espradio_mutex_unlock,
    ._queue_create = espradio_queue_create,
    ._queue_delete = espradio_queue_delete,
    ._queue_send = espradio_queue_send,
    ._queue_send_from_isr = espradio_queue_send_from_isr,
    ._queue_send_to_back = espradio_queue_send_to_back,
    ._queue_send_to_front = espradio_queue_send_to_front,
    ._queue_recv = espradio_queue_recv,
    ._queue_msg_waiting = espradio_queue_msg_waiting,
    ._event_group_create = espradio_event_group_create,
    ._event_group_delete = espradio_event_group_delete,
    ._event_group_set_bits = espradio_event_group_set_bits,
    ._event_group_clear_bits = espradio_event_group_clear_bits,
    ._event_group_wait_bits = espradio_event_group_wait_bits,
    ._task_create_pinned_to_core = espradio_task_create_pinned_to_core,
    ._task_create = espradio_task_create,
    ._task_delete = espradio_task_delete,
    ._task_delay = espradio_task_delay,
    ._task_ms_to_tick = espradio_task_ms_to_tick,
    ._task_get_current_task = espradio_task_get_current_task,
    ._task_get_max_priority = espradio_task_get_max_priority,
    ._malloc = espradio_malloc,
    ._free = espradio_free,
    ._event_post = espradio_event_post,
    ._get_free_heap_size = espradio_get_free_heap_size,
    ._rand = espradio_rand,
    ._dport_access_stall_other_cpu_start_wrap = espradio_dport_access_stall_other_cpu_start_wrap,
    ._dport_access_stall_other_cpu_end_wrap = espradio_dport_access_stall_other_cpu_end_wrap,
    ._wifi_apb80m_request = espradio_wifi_apb80m_request,
    ._wifi_apb80m_release = espradio_wifi_apb80m_release,
    ._phy_disable = espradio_phy_disable,
    ._phy_enable = espradio_phy_enable,
    ._phy_update_country_info = espradio_phy_update_country_info,
    ._read_mac = espradio_read_mac,
    ._timer_arm = espradio_timer_arm,
    ._timer_disarm = espradio_timer_disarm,
    ._timer_done = espradio_timer_done,
    ._timer_setfn = espradio_timer_setfn,
    ._timer_arm_us = espradio_timer_arm_us,
    ._wifi_reset_mac = espradio_wifi_reset_mac,
    ._wifi_clock_enable = espradio_wifi_clock_enable,
    ._wifi_clock_disable = espradio_wifi_clock_disable,
    ._wifi_rtc_enable_iso = espradio_wifi_rtc_enable_iso,
    ._wifi_rtc_disable_iso = espradio_wifi_rtc_disable_iso,
    ._esp_timer_get_time = espradio_esp_timer_get_time,
    ._nvs_set_i8 = espradio_nvs_set_i8,
    ._nvs_get_i8 = espradio_nvs_get_i8,
    ._nvs_set_u8 = espradio_nvs_set_u8,
    ._nvs_get_u8 = espradio_nvs_get_u8,
    ._nvs_set_u16 = espradio_nvs_set_u16,
    ._nvs_get_u16 = espradio_nvs_get_u16,
    ._nvs_open = espradio_nvs_open,
    ._nvs_close = espradio_nvs_close,
    ._nvs_commit = espradio_nvs_commit,
    ._nvs_set_blob = espradio_nvs_set_blob,
    ._nvs_get_blob = espradio_nvs_get_blob,
    ._nvs_erase_key = espradio_nvs_erase_key,
    ._get_random = espradio_get_random,
    ._get_time = espradio_get_time,
    ._random = espradio_random,
    ._slowclk_cal_get = espradio_slowclk_cal_get,
    ._log_write = espradio_log_write,
    ._log_writev = espradio_log_writev,
    ._log_timestamp = espradio_log_timestamp,
    ._malloc_internal = espradio_malloc_internal,
    ._realloc_internal = espradio_realloc_internal,
    ._calloc_internal = espradio_calloc_internal,
    ._zalloc_internal = espradio_zalloc_internal,
    ._wifi_malloc = espradio_wifi_malloc,
    ._wifi_realloc = espradio_wifi_realloc,
    ._wifi_calloc = espradio_wifi_calloc,
    ._wifi_zalloc = espradio_wifi_zalloc,
    ._wifi_create_queue = espradio_wifi_create_queue,
    ._wifi_delete_queue = espradio_wifi_delete_queue,
    ._coex_init = espradio_coex_init,
    ._coex_deinit = espradio_coex_deinit,
    ._coex_enable = espradio_coex_enable,
    ._coex_disable = espradio_coex_disable,
    ._coex_status_get = espradio_coex_status_get,
    ._coex_condition_set = espradio_coex_condition_set,
    ._coex_wifi_request = espradio_coex_wifi_request,
    ._coex_wifi_release = espradio_coex_wifi_release,
    ._coex_wifi_channel_set = espradio_coex_wifi_channel_set,
    ._coex_event_duration_get = espradio_coex_event_duration_get,
    ._coex_pti_get = espradio_coex_pti_get,
    ._coex_schm_status_bit_clear = espradio_coex_schm_status_bit_clear,
    ._coex_schm_status_bit_set = espradio_coex_schm_status_bit_set,
    ._coex_schm_interval_set = espradio_coex_schm_interval_set,
    ._coex_schm_interval_get = espradio_coex_schm_interval_get,
    ._coex_schm_curr_period_get = espradio_coex_schm_curr_period_get,
    ._coex_schm_curr_phase_get = espradio_coex_schm_curr_phase_get,
    ._coex_schm_process_restart = espradio_coex_schm_process_restart,
    ._coex_schm_register_cb = espradio_coex_schm_register_cb,
    ._coex_register_start_cb = espradio_coex_register_start_cb,
    ._magic = ESP_WIFI_OS_ADAPTER_MAGIC,
};
