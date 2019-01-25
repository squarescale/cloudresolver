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
}

type CloudResolver interface {
	Resolve(string, map[string]interface{}) ([]Host, error)
}

var Resolvers []CloudResolver

func register(provider CloudResolver) {
	Resolvers = append(Resolvers, provider)
}
