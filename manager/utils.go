package manager

import (
	"errors"
	"fmt"

	"github.com/arkbriar/ss-mgr/manager/protocol"
)

func constructProtocolServiceList(srvs ...*shadowsocksService) *protocol.ServiceList {
	services := make([]*protocol.ShadowsocksService, 0, len(srvs))
	for _, srv := range srvs {
		services = append(services, &protocol.ShadowsocksService{
			Port:     srv.Port,
			Password: srv.Password,
		})
	}
	return &protocol.ServiceList{
		Services: services,
	}
}

func constructServiceList(srvList *protocol.ServiceList) []*shadowsocksService {
	srvs := make([]*shadowsocksService, 0, len(srvList.GetServices()))
	for _, pbsrv := range srvList.GetServices() {
		srvs = append(srvs, &shadowsocksService{
			UserId:   pbsrv.GetUserId(),
			Port:     pbsrv.GetPort(),
			Password: pbsrv.GetPassword(),
		})
	}
	return srvs
}

func compareLists(required, current *protocol.ServiceList) (diff []*shadowsocksService) {
	diff = make([]*shadowsocksService, 0, 1)
	for _, a := range required.GetServices() {
		for _, b := range current.GetServices() {
			if a.GetUserId() == b.GetUserId() && a.GetPassword() == b.GetPassword() {
				break
			}
		}
		diff = append(diff, &shadowsocksService{
			UserId:   a.GetUserId(),
			Port:     -1,
			Password: a.GetPassword(),
		})
	}
	return diff
}

func constructErrorFromDifferenceServiceList(diff []*shadowsocksService) error {
	if diff == nil || len(diff) == 0 {
		return nil
	}
	var errMsg string
	if len(diff) == 1 {
		errMsg = fmt.Sprintf("There is 1 service not allocated (user: password):")
	} else {
		errMsg = fmt.Sprintf("There are %d services not allocated (user: password):", len(diff))
	}
	for _, srv := range diff {
		errMsg += fmt.Sprintf("\n  %s: %s", srv.UserId, srv.Password)
	}
	return errors.New(errMsg)
}
