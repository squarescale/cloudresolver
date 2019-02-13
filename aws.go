package cloudresolver

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

type AwsResolver struct{}

func init() {
	register("aws", new(AwsResolver))
}

func (r AwsResolver) Resolve(name string, config map[string]interface{}) ([]Host, error) {
	log.Printf("in aws: %+v\n", Resolvers["do"])
	log.Info("starting aws")
	// on EC2, linux, kvm, this file contains "Amazon EC2"
	vendor, _ := ioutil.ReadFile("/sys/devices/virtual/dmi/id/sys_vendor")
	if !bytes.Equal([]byte("Amazon EC2"), vendor) {
		// disable ec2 metadata role when not on ec2
		os.Setenv("AWS_EC2_METADATA_DISABLED", "true")
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := ec2.New(sess)

	filters := []*ec2.Filter{
		// builtin filter, the purpose is connecting to hosts, they need to be running
		&ec2.Filter{
			Name:   aws.String("instance-state-name"),
			Values: []*string{aws.String("running")},
		},
	}

	f := &ec2.Filter{
		Name:   aws.String("tag:Name"),
		Values: []*string{aws.String(name)},
	}
	filters = append(filters, f)

	params := &ec2.DescribeInstancesInput{
		Filters: filters,
	}

	resp, err := svc.DescribeInstances(params)
	if err != nil {
		return []Host{}, err
	}

	hosts := []Host{}
	for idx, _ := range resp.Reservations {
		for _, inst := range resp.Reservations[idx].Instances {
			h := Host{
				Provider:    "aws",
				Zone:        *inst.Placement.AvailabilityZone,
				Id:          *inst.InstanceId,
				PrivateIpv4: *inst.PrivateIpAddress,
				PrivateName: *inst.PrivateDnsName,
				PublicName:  *inst.PublicDnsName,
				Public:      *inst.PublicDnsName,
			}
			hosts = append(hosts, h)
		}
	}

	return hosts, nil
}
