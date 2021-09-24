#include <stdint.h>
#include <stddef.h>
#include <stdio.h>
#include "espnet.h"

// OS adapter functions.
// See: esp-idf/components/esp_wifi/include/esp_private/wifi_os_adapter.h
wifi_osi_funcs_t g_wifi_osi_funcs = {
    ._version = ESP_WIFI_OS_ADAPTER_VERSION,
    // TODO: define these functions.
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
