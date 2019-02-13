package cloudresolver

import log "github.com/sirupsen/logrus"

type LocalResolver struct{}

func init() {
	register("local", new(LocalResolver))
}

func (r LocalResolver) Resolve(name string, config map[string]interface{}) ([]Host, error) {
	log.Info("starting local")
	h := Host{Provider: "local", PrivateIpv4: "127.0.0.1"}
	return []Host{h}, nil
}
