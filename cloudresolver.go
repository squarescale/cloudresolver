package cloudresolver

type Host struct {
	Provider    string
	Region      string
	Zone        string
	Id          string
	PrivateIpv4 string
	PrivateIpv6 string
	PrivateName string
	PublicIpv4  string
	PublicIpv6  string
	PublicName  string
	Private     string // Either ip or name, the one accessible on current cloud
	Public      string
}

type CloudResolver interface {
	Resolve(string, map[string]interface{}) ([]Host, error)
}

var Resolvers map[string]CloudResolver

func register(name string, provider CloudResolver) {
	if len(Resolvers) == 0 {
		Resolvers = make(map[string]CloudResolver)
	}
	Resolvers[name] = provider
}
