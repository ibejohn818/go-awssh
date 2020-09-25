package api

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/aws/aws-sdk-go/service/ec2instanceconnect"
	"github.com/aws/aws-sdk-go/service/ec2instanceconnect/ec2instanceconnectiface"
)

type Ec2ConnClient struct {
	Sdk ec2instanceconnectiface.EC2InstanceConnectAPI
}

func NewEc2ConnClient(ops ...func(*Ec2ConnClient)) *Ec2ConnClient {
	c := Ec2ConnClient{}

	if len(ops) == 1 {
		ops[0](&c)
	} else {
		c.Sdk = ec2instanceconnect.New(defaultSession(), defaultConfig())
	}

	return &c
}

type Ec2ConnPayload struct {
	User       string
	PubKeyPath string
	Instance   Ec2Instance
}

/*
func NewEc2ConnPayload() *Ec2ConnPayload {

}
*/

func (client *Ec2ConnClient) SendPublicKey(payload *Ec2ConnPayload) {

	pubKey, err := ioutil.ReadFile(payload.PubKeyPath)

	if err != nil {
		log.Fatal("Unable to open public key")
	}

	pubKeyStr := string(pubKey)

	inp := ec2instanceconnect.SendSSHPublicKeyInput{
		AvailabilityZone: &payload.Instance.AZ,
		InstanceId:       &payload.Instance.InstanceID,
		InstanceOSUser:   &payload.User,
		SSHPublicKey:     &pubKeyStr,
	}

	//spew.Dump(inp)

	res, err := client.Sdk.SendSSHPublicKey(&inp)

	if err != nil || !*res.Success {
		fmt.Println(err)
		log.Fatal("Error sending key to instance")
	}
}
