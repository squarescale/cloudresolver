package cloudresolver

type LocalResolver struct{}

func init() {
	register("local", new(LocalResolver))
}

func (r LocalResolver) Resolve(name string, config map[string]interface{}) ([]Host, error) {
	h := Host{
		InstanceName: "localhost",
		Provider:     "local",
		Zone:         "local",
		Region:       "local",
		PrivateIpv4:  "127.0.0.1",
		PublicIpv4:   "127.0.0.1",
		PrivateIpv6:  "::1",
		PublicIpv6:   "::1",
		PublicName:   "localhost",
		PrivateName:  "localhost",
		Public:       "localhost",
		Private:      "localhost",
	}
	return []Host{h}, nil
}
