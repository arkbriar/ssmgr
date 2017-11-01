#pragma once

#ifndef __SSMGR_COLLECTOR_CONFIG_H__
#define __SSMGR_COLLECTOR_CONFIG_H__

#include "common.h"
#include <memory>

namespace ssmgr {

struct ShadowsocksLibevConfig {
    string remote_host;
    uint16_t remote_port;
    string local_host;
    uint16_t local_port;
};

class CollectorConfig : public ShadowsocksLibevConfig {
public:
    // TODO plugin opts

    static std::shared_ptr<CollectorConfig> read_config_from_env();

protected:
    CollectorConfig() {}

private:
    void resolv_from(const string& opts);
};

}  // namespace ssmgr

#endif
