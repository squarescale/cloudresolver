package cloudresolver

import (
	"bytes"
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type AwsResolver struct{}

func init() {
	register("aws", new(AwsResolver))
}

func (r AwsResolver) Resolve(name string, config map[string]interface{}) ([]Host, error) {
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

	// Primary filter on instance state (running) and Name TAG
	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("instance-state-name"),
				Values: []*string{aws.String("running")},
			},
			{
				Name:   aws.String("tag:Name"),
				Values: []*string{aws.String(name)},
			},
		},
	}
	// Do the lookup
	resp, err := svc.DescribeInstances(params)
	if err != nil {
		// Return the error to caller
		return []Host{}, err
	}

	// No match found, try secondary filter on instance state (running) and ID
	if len(resp.Reservations) == 0 {
		params := &ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("instance-state-name"),
					Values: []*string{aws.String("running")},
				},
				{
					Name:   aws.String("instance-id"),
					Values: []*string{aws.String(name)},
				},
			},
		}
		// Do the lookup
		resp, err = svc.DescribeInstances(params)
		if err != nil {
			// Return the error to caller
			return []Host{}, err
		}
	}

	// No match found, try tertiary filter on instance state (running) and private IP
	if len(resp.Reservations) == 0 {
		params := &ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("instance-state-name"),
					Values: []*string{aws.String("running")},
				},
				{
					Name:   aws.String("private-ip-address"),
					Values: []*string{aws.String(name)},
				},
			},
		}
		// Do the lookup
		resp, err = svc.DescribeInstances(params)
		if err != nil {
			// Return the error to caller
			return []Host{}, err
		}
	}

	hosts := []Host{}
	for idx, _ := range resp.Reservations {
		for _, inst := range resp.Reservations[idx].Instances {
			iname := ""
			for _, tag := range inst.Tags {
				if *tag.Key == "Name" {
					iname = *tag.Value
				}
			}
			h := Host{
				InstanceName: iname,
				Provider:     "aws",
				Region:       *sess.Config.Region,
				Zone:         *inst.Placement.AvailabilityZone,
				Id:           *inst.InstanceId,
				PrivateIpv4:  *inst.PrivateIpAddress,
				PrivateName:  *inst.PrivateDnsName,
				PublicName:   *inst.PublicDnsName,
				Private:      *inst.PrivateIpAddress,
				Public:       *inst.PublicDnsName,
			}
			hosts = append(hosts, h)
		}
	}

	return hosts, nil
}
