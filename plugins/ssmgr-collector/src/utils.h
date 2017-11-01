#pragma once

#ifndef __SSMGR_COLLECTOR_UTILS_H__
#define __SSMGR_COLLECTOR_UTILS_H__

#include "common.h"

/* Util functions for loading environment variable into specified variables.
 * Return true when correspoding environment variable is set and valid, and false
 * otherwise.
 *
 * @env_var: environment variable name, e.x. "REMOTE_HOST"
 * @env_val: variable to set */
bool load_env(const string& env_var, string& env_val);
bool load_env(const string& env_var, uint16_t& env_val);

#endif
