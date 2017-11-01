#include "utils.h"

bool load_env(const string& env_var, string& env_val) {
    char* env_val_c = std::getenv(env_var.c_str());
    if (env_val_c) {
        env_val = string(env_val_c);
        return !env_val.empty();
    }
    return false;
}

bool load_env(const string& env_var, uint16_t& env_val) {
    char* env_val_c = std::getenv(env_var.c_str());
    if (env_val_c) {
        env_val = ::atoi(env_val_c);
        return env_val != 0;
    }
    return false;
}
