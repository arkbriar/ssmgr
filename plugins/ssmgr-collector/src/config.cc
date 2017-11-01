#include "config.h"
#include "utils.h"

#include <cstdlib>

namespace ssmgr {

static std::shared_ptr<CollectorConfig> global = nullptr;

std::shared_ptr<CollectorConfig> CollectorConfig::read_config_from_env() {
    // if global set, return global
    // otherwise initialize global
    if (!global) {
        global = std::make_shared<CollectorConfig>();
        CHECK(load_env("SS_REMOTE_HOST", global->remote_host))
            << "Environment variable \"SS_REMOTE_HOST\" is not set or illegal!";
        CHECK(load_env("SS_REMOTE_PORT", global->remote_port))
            << "Environment variable \"SS_REMOTE_PORT\" is not set or illegal!";
        CHECK(load_env("SS_LOCAL_HOST", global->local_host))
            << "Environment variable \"SS_LOCAL_HOST\" is not set or illegal!";
        CHECK(load_env("SS_LOCAL_PORT", global->local_port))
            << "Environment variable \"SS_LOCAL_PORT\" is not set or illegal!";

        string plugin_opts;
        if (load_env("SS_PLUGIN_OPTIONS", plugin_opts)) {
            VLOG(5) << "Plugin options: " << plugin_opts;
            global->resolv_from(plugin_opts);
        }
    }
    return global;
}

void CollectorConfig::resolv_from(const string& plugin_opts) {
    //
}

}  // namespace ssmgr
