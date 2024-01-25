typedef int _lock_t;

#define SOC_COEX_HW_PTI 1

#include "espidf_types.h"
#include "esp_private/wifi.h"
#include "esp_wpa.h"
#include "esp_phy_init.h"
#include "phy.h"
#include "esp_timer.h"

#if !defined(CONFIG_IDF_TARGET_ESP32S2)
#include "esp_bt.h"
#include "esp_coexist_internal.h"
#include "esp_coexist_adapter.h"
#endif

#include "esp_now.h"
