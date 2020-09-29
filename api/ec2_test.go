package api_test

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/davecgh/go-spew/spew"
	"github.com/ibejohn818/awssh/api"
)

type mockSdkGood struct {
	ec2iface.EC2API
}

func (sdk mockSdkGood) DescribeAddresses(inp *ec2.DescribeAddressesInput) (*ec2.DescribeAddressesOutput, error) {
	file, _ := ioutil.ReadFile(filepath.Join("testdata", "describeAddressesGood.json"))
	spew.Dump(string(file))
	out := ec2.DescribeAddressesOutput{}
	_ = json.Unmarshal([]byte(file), &out)
	return &out, nil
}

func (sdk mockSdkGood) DescribeInstances(inp *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	file, _ := ioutil.ReadFile(filepath.Join("testdata", "describeInstancesGood.json"))
	out := ec2.DescribeInstancesOutput{}
	_ = json.Unmarshal([]byte(file), &out)
	return &out, nil
}

func TestGetEips(t *testing.T) {

	ec2Client := api.NewEc2Client(func(ops *api.Ec2Client) {
		ops.Sdk = mockSdkGood{}
	})

	res := ec2Client.GetEips()

	spew.Dump(res)

}
