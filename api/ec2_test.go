package api_test

import (
	"github.com/aws/aws-sdk-go/service/ec2"
)

type MockSdk struct{}

func (sdk *MockSdk) DescribeAddresses(inp *ec2.DescribeAddressesInput) (*ec2.DescribeAddressesOutput, error) {
	res := ec2.DescribeAddressesOutput{}

	return &res, nil
}
