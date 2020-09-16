package compute

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/davecgh/go-spew/spew"
)

type Ec2Instance struct {
	InstanceID string
	Name       string
	Ip         string
	PrivateIp  string
}

func GetInstances(sdk Ec2Sdk) []Ec2Instance {
	ii := make([]Ec2Instance, 0, 1)

	inp := &ec2.DescribeInstancesInput{}

	res, err := sdk.Client.DescribeInstances(inp)

	if err != nil {
		fmt.Println(err)
	}

	for _, v := range res.Reservations {
		for _, vv := range v.Instances {

			if ip := vv.PublicIpAddress; ip != nil {

				i := &Ec2Instance{
					Ip: *ip,
					// Name:      name,
					PrivateIp: *vv.PrivateIpAddress,
				}

				// if ec2Coll.HasEip(i) {
				// 	i.HasEip = true
				// }

				ii = append(ii, *i)
			}
		}
	}

	// slice.Sort(ec2Coll.Instances, func(i, j int) bool {
	// 	return ec2Coll.Instances[i].Name < ec2Coll.Instances[j].Name
	// })
	// query, err := inst.Client.
	return ii
}

func parseTags(tags []ec2.Tag) {
	name := ""

	for _, tt := range tags {
		if *tt.Key == "Name" {
			name = *tt.Value
		}
	}
	spew.Dump(name)
}
