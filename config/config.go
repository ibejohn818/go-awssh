package config

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/spf13/viper"
)

type AwsshConf struct {
	Region string
	VpConf *viper.Viper
}

func (conf *AwsshConf) GetAwsConf() *aws.Config {
	ac := aws.NewConfig().WithRegion(conf.VpConf.GetString("region"))
	return ac
}
