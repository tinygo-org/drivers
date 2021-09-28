#include <stdbool.h>

#define XTOS_SET_INTLEVEL(level) 0
#define XTOS_RESTORE_INTLEVEL(level) 0

#define RSR(reg, at)         asm volatile ("rsr %0, %1" : "=r" (at) : "i" (reg))

struct _reent;

#include "esp_private/wifi_os_adapter.h"

extern void wifi_init_default(void* cfg);
extern wifi_osi_funcs_t g_wifi_osi_funcs;
extern void wifi_osi_lend_memory( void* ptr, uint32_t size );
