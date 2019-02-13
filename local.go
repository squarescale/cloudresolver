package cloudresolver

import log "github.com/sirupsen/logrus"

type LocalResolver struct{}

func init() {
	register("local", new(LocalResolver))
}

func (r LocalResolver) Resolve(name string, config map[string]interface{}) ([]Host, error) {
	log.Info("starting local")
	h := Host{
		Provider:    "local",
		Zone:        "local",
		Region:      "local",
		PrivateIpv4: "127.0.0.1",
		PublicIpv4:  "127.0.0.1",
		PrivateIpv6: "::1",
		PublicIpv6:  "::1",
		PublicName:  "localhost",
		PrivateName: "localhost",
		Public:      "localhost",
		Private:     "localhost",
	}
	return []Host{h}, nil
}
