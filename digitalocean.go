package cloudresolver

import "gopkg.in/yaml.v2"
import (
	"context"
	"fmt"
	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
	"io/ioutil"
	"os"
	"path/filepath"
)

type TokenSource struct {
	AccessToken string
}

func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

type DigitalOceanResolver struct{}

func init() {
	register("do", new(DigitalOceanResolver))
}

type YamlToken struct {
	token string `yaml:"access-token"`
}

func (r DigitalOceanResolver) Resolve(name string, config map[string]interface{}) ([]Host, error) {
	var config_home string
	if xdgPath := os.Getenv("XDG_CONFIG_HOME"); xdgPath != "" {
		config_home = filepath.Join(xdgPath, "doctl")
	} else {
		config_home = os.Getenv("HOME")
	}

	doctlcfg := filepath.Join(config_home, ".config", "doctl", "config.yaml")

	var docfg map[string]interface{}
	yf, err := ioutil.ReadFile(doctlcfg)
	if err != nil {
		return []Host{}, err
	}

	err = yaml.Unmarshal(yf, &docfg)
	if err != nil {
		fmt.Printf("Unmarshal: %v", err)
	}

	fmt.Printf("%+v\n", docfg)

	tokenSource := &TokenSource{
		AccessToken: docfg["access-token"].(string),
	}

	oauthClient := oauth2.NewClient(context.Background(), tokenSource)
	client := godo.NewClient(oauthClient)
	droplets, _, err := client.Droplets.List(context.Background(), nil)
	if err != nil {
		return []Host{}, err
	}

	hosts := []Host{}
	for _, droplet := range droplets {

		if droplet.Name == name {
			publicIpv4, _ := droplet.PublicIPv4()
			privateIpv4, _ := droplet.PrivateIPv4()
			publicIpv6, _ := droplet.PublicIPv6()
			h := Host{
				Provider:    "digitalocean",
				Region:      droplet.Region.Slug,
				Id:          fmt.Sprintf("%v", droplet.ID),
				PublicIpv4:  publicIpv4,
				PrivateIpv4: privateIpv4,
				PublicIpv6:  publicIpv6,
				Private:     privateIpv4,
				Public:      publicIpv4,
			}
			hosts = append(hosts, h)
		}
	}

	return hosts, nil
}
