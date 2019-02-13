package cloudresolver

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
	"io/ioutil"
	"net/http"
	"os"
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
	log.Printf("v: %+v\n", v)

	if err != nil {
		log.Printf("v: %+v\n", v)
		log.Printf("err: %+v\n", err)
		return []Host{}, err
	}

	client, err := client(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	if err != nil {
		log.Printf("%+v\n", err)
		return []Host{}, err
	}

	creds, err := google.FindDefaultCredentials(context.TODO(), "")
	if err != nil {
		log.Printf("%+v\n", err)
	}

	service, err := compute.New(client)
	if err != nil {
		log.Printf("%+v\n", err)
		return []Host{}, err
	}

	res, err := service.Instances.Get(creds.ProjectID, v.GetString("providers.gce.zone"), name).Do()
	if err != nil {
		log.Printf("%+v\n", err)
		return []Host{}, err
	}

	if len(res.NetworkInterfaces) == 0 || res.NetworkInterfaces[0].NetworkIP == "" {
		return []Host{}, err
	}

	h := Host{
		Provider:    "gce",
		Zone:        res.Zone,
		Id:          fmt.Sprintf("%s", res.Id),
		PrivateIpv4: res.NetworkInterfaces[0].NetworkIP,
		//PrivateName: *inst.PrivateDnsName,
		//PublicName:  *inst.PublicDnsName,
	}

	return []Host{h}, err

}
