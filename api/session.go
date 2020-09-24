package api

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

var (
	Region  string
	Profile string
)

func init() {
	Region = "us-west-2"
	Profile = "default"
}

func defaultSession() *session.Session {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	return sess
}

func defaultConfig() *aws.Config {
	conf := aws.NewConfig().WithRegion(Region)
	return conf
}
