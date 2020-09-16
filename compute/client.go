package compute

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

type Ec2Sdk struct {
	Client ec2iface.EC2API
}

func NewEc2Sdk(ops ...func(*Ec2Sdk)) Ec2Sdk {
	sdk := Ec2Sdk{}

	if len(ops) == 1 {
		ops[0](&sdk)
	} else {
		sess := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}))
		sdk.Client = ec2.New(sess)
	}

	return sdk
}
