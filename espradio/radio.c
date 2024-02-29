//go:build esp32c3

#include "sdkconfig.h"
#include "include.h"
#include <stdarg.h>

// defined in osi.c
extern wifi_osi_funcs_t espradio_osi_funcs;

wifi_init_config_t wifi_config = {
    .osi_funcs = &espradio_osi_funcs,
    .wpa_crypto_funcs = {
        .size = sizeof(wpa_crypto_funcs_t),
        .version = ESP_WIFI_CRYPTO_VERSION,
        // TODO: fill in these functions
    },
    .static_rx_buf_num = CONFIG_ESP_WIFI_STATIC_RX_BUFFER_NUM,
    .dynamic_rx_buf_num = CONFIG_ESP_WIFI_DYNAMIC_RX_BUFFER_NUM,
    .tx_buf_type = CONFIG_ESP_WIFI_TX_BUFFER_TYPE,
    .static_tx_buf_num = WIFI_STATIC_TX_BUFFER_NUM,
    .dynamic_tx_buf_num = WIFI_DYNAMIC_TX_BUFFER_NUM,
    .rx_mgmt_buf_type = CONFIG_ESP_WIFI_DYNAMIC_RX_MGMT_BUF,
    .rx_mgmt_buf_num = WIFI_RX_MGMT_BUF_NUM_DEF,
    .cache_tx_buf_num = WIFI_CACHE_TX_BUFFER_NUM,
    .csi_enable = WIFI_CSI_ENABLED,
    .ampdu_rx_enable = WIFI_AMPDU_RX_ENABLED,
    .ampdu_tx_enable = WIFI_AMPDU_TX_ENABLED,
    .amsdu_tx_enable = WIFI_AMSDU_TX_ENABLED,
    .nvs_enable = 0, // currently unsupported
    .nano_enable = WIFI_NANO_FORMAT_ENABLED,
    .rx_ba_win = WIFI_DEFAULT_RX_BA_WIN,
    .wifi_task_core_id = WIFI_TASK_CORE_ID,
    .beacon_max_len = WIFI_SOFTAP_BEACON_MAX_LEN,
    .mgmt_sbuf_num = WIFI_MGMT_SBUF_NUM,
    .feature_caps = WIFI_FEATURE_CAPS,
    .sta_disconnected_pm = WIFI_STA_DISCONNECTED_PM_ENABLED,
    .espnow_max_encrypt_num = CONFIG_ESP_WIFI_ESPNOW_MAX_ENCRYPT_NUM,
    .magic = WIFI_INIT_CONFIG_MAGIC
};

static char evt;
esp_event_base_t const WIFI_EVENT = &evt;

void net80211_printf(const char *format, ...) {
    va_list args;
    va_start(args, format);
    printf("espradio net80211: ");
    vprintf(format, args);
    va_end(args);
}

void phy_printf(const char *format, ...) {
    va_list args;
    va_start(args, format);
    printf("espradio phy: ");
    vprintf(format, args);
    va_end(args);
}

void pp_printf(const char *format, ...) {
    va_list args;
    va_start(args, format);
    printf("espradio pp: ");
    vprintf(format, args);
    va_end(args);
}
