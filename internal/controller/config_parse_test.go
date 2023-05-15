package controller

import (
	"fmt"
	"testing"

	fcorev1 "github.com/Faione/frp-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIniParser(t *testing.T) {

	sp := &fcorev1.ServiceProxy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
		Spec: fcorev1.ServiceProxySpec{
			Proxies: []fcorev1.ProxyEntry{
				{
					Name:       "example_server",
					LocalIp:    "echoserver-service",
					LocalPort:  "8080",
					RemotePort: "8888",
				},
			},
		},
	}

	frps := &fcorev1.FrpService{
		Spec: fcorev1.FrpServiceSpec{
			Address: "frpservice-sample",
			Port:    17000,
			Token:   "fanghao@lei",
		},
	}

	configMap := defaultConfigMap(sp, frps)

	fmt.Println(configMap.Data["frpc.ini"])
}
