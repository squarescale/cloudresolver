package cloudresolver

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	compute "google.golang.org/api/compute/v1"

	"github.com/spf13/viper"
)

type GceResolver struct {
}

func init() {
	register("gce", new(GceResolver))
}

func client(path string) (*http.Client, error) {
	if path == "" {
		return google.DefaultClient(oauth2.NoContext, compute.ComputeScope)
	}

	key, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	jwtConfig, err := google.JWTConfigFromJSON(key, compute.ComputeScope)
	if err != nil {
		return nil, err
	}

	return jwtConfig.Client(oauth2.NoContext), nil
}

func (r GceResolver) Resolve(name string, config map[string]interface{}) ([]Host, error) {
	v := viper.New()
	err := v.MergeConfigMap(config)

	if err != nil {
		return []Host{}, err
	}

	client, err := client(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	if err != nil {
		return []Host{}, err
	}

	creds, err := google.FindDefaultCredentials(context.TODO(), "")
	if err != nil {
		return []Host{}, err
	}

	service, err := compute.New(client)
	if err != nil {
		return []Host{}, err
	}

	res, err := service.Instances.Get(creds.ProjectID, v.GetString("providers.gce.zone"), name).Do()
	if err != nil {
		return []Host{}, err
	}

	if len(res.NetworkInterfaces) == 0 || res.NetworkInterfaces[0].NetworkIP == "" {
		return []Host{}, err
	}

	h := Host{
		InstanceName: res.Name,
		Provider:     "gce",
		Zone:         v.GetString("providers.gce.zone"),
		Id:           fmt.Sprintf("%d", res.Id),
		PublicIpv4:   res.NetworkInterfaces[0].AccessConfigs[0].NatIP,
		Public:       res.NetworkInterfaces[0].AccessConfigs[0].NatIP,
		PrivateIpv4:  res.NetworkInterfaces[0].NetworkIP,
		Private:      res.NetworkInterfaces[0].NetworkIP,
	}

	return []Host{h}, err
}
