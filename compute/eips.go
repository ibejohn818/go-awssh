package compute

import (
	"github.com/aws/aws-sdk-go/service/ec2"
)

type ElasticIp struct {
	Ip           *string
	AllocationID *string
}

func LoadEips(sdk Ec2Sdk) []ElasticIp {

	// sess := session.Must(session.NewSession(&aws.Config{
	// 	Region: aws.String(endpoints.UsWest2RegionID),
	// }))

	inp := &ec2.DescribeAddressesInput{}

	res, err := sdk.Client.DescribeAddresses(inp)

	if err != nil {
		panic(err)
	}

	ips := make([]ElasticIp, 0, 1)

	for _, v := range res.Addresses {
		ips = append(ips, ElasticIp{
			Ip:           v.PublicIp,
			AllocationID: v.AllocationId,
		})
	}

	return ips
}
