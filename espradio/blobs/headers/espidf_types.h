// Copyright 2020 Espressif Systems (Shanghai) PTE LTD
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

#ifndef _ESPIDF_TYPES_H_
#define _ESPIDF_TYPES_H_

#include <stdint.h>

#define IDF_DEPRECATED(_s)

enum esp_log_level_e
{
  ESP_LOG_NONE,
  ESP_LOG_ERROR,
  ESP_LOG_WARN,
  ESP_LOG_INFO,
  ESP_LOG_DEBUG,
  ESP_LOG_VERBOSE
};

typedef uint32_t        TickType_t;
typedef uint32_t        UBaseType_t;
typedef int32_t         BaseType_t;

typedef void*           QueueHandle_t;

typedef void*           esp_netif_t;
typedef void*           esp_netif_inherent_config_t;

struct ets_timer
{
  struct timer_adpt *next;
  uint32_t expire;
  uint32_t period;
  void (*func)(void *priv);
  void *priv;
};

#endif /* _ESPIDF_TYPES_H_ */
