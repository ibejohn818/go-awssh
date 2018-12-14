package api

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/bradfitz/slice"
	"github.com/ibejohn818/awssh/config"
)

var ec2Client = ec2.New(session.New())

func try() {

}

type Ec2Instance struct {
	Name      string
	Ip        string
	PrivateIp string
	HasEip    bool
}

func (e *Ec2Instance) GetTplMap() map[string]string {

	m := make(map[string]string)

	m["Name"] = e.Name

	m["Ip"] = fmt.Sprintf("%15v", e.Ip)

	m["PrivateIp"] = fmt.Sprintf("%15v", e.PrivateIp)

	if e.HasEip {
		m["HasEip"] = "✓"
	} else {
		m["HasEip"] = " "
	}

	return m
}

func (e *Ec2Instance) GetLine() string {

	tmpl, err := template.New("ListServers").Parse("{{ .HasEip }} [{{ .Ip }}]: {{ .Name }}")

	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer

	m := e.GetTplMap()

	tmpl.Execute(&buf, m)

	return buf.String()

}

func (e *Ec2Instance) GetLinePrivate() string {

	tmpl, err := template.New("ListServers").Parse("[{{ .PrivateIp }}]: {{ .Name }}")

	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer

	m := e.GetTplMap()

	tmpl.Execute(&buf, m)

	return buf.String()

}

func (e *Ec2Instance) Match(args []string) bool {

	sname := strings.ToLower(e.Name)

	for _, v := range args {

		seed := strings.ToLower(v)

		// check for ^ operator
		if m, _ := regexp.MatchString("^\\^", seed); m {

			seed = strings.Replace(seed, "^", "", 1)

			if m, _ := regexp.MatchString(seed, sname); m {
				return false
			}

		} else {
			if m, _ := regexp.MatchString(seed, sname); !m {
				return false
			}
		}

	}

	return true
}

// Eip is an elastic ip from the AWS API
type Eip struct {
	Ip string
}

// Ec2Collection holds structs of Ec2Instance's and Eip's queried from AWS API
type Ec2Collection struct {
	Instances []Ec2Instance
	EIPs      []Eip
}

// ListServers prints our a formatted list of the
// instances slice
func (col *Ec2Collection) ListServers() {

	tmpl, err := template.New("ListServers").Parse("{{ .HasEip }} [{{ .Ip }}]: {{ .Name }}\n")

	if err != nil {
		panic(err)
	}

	for _, inst := range col.Instances {
		m := inst.GetTplMap()
		tmpl.Execute(os.Stdout, m)
	}
}

func (col *Ec2Collection) LoadEips(conf *config.AwsshConf) {

	ec := ec2.New(session.New(), conf.GetAwsConf())

	inp := &ec2.DescribeAddressesInput{}

	res, err := ec.DescribeAddresses(inp)

	if err != nil {
		panic(err)
	}

	for _, v := range res.Addresses {
		col.EIPs = append(col.EIPs, Eip{
			Ip: *v.PublicIp,
		})
	}

}

// HasEip will check if the Ec2Instance public ip is a ElasticIp
func (col *Ec2Collection) HasEip(e *Ec2Instance) bool {

	for _, v := range col.EIPs {
		if v.Ip == e.Ip {
			return true
		}
	}

	return false
}

// GetServers is a constructor for Ec2Collection and loads
// instances and EIPs slice
func GetServers(conf *config.AwsshConf) *Ec2Collection {

	ec := ec2.New(session.New(), conf.GetAwsConf())

	ec2Coll := &Ec2Collection{}

	ec2Coll.LoadEips(conf)

	inp := &ec2.DescribeInstancesInput{}

	res, err := ec.DescribeInstances(inp)

	if err != nil {
		fmt.Println(err)
	}

	for _, v := range res.Reservations {
		for _, vv := range v.Instances {
			name := ""
			if tags := vv.Tags; tags != nil {

				for _, tt := range tags {
					if *tt.Key == "Name" {
						name = *tt.Value
					}
				}
			}
			if ip := vv.PublicIpAddress; ip != nil {
				i := &Ec2Instance{
					Ip:        *ip,
					Name:      name,
					PrivateIp: *vv.PrivateIpAddress,
				}

				if ec2Coll.HasEip(i) {
					i.HasEip = true
				}

				ec2Coll.Instances = append(ec2Coll.Instances, *i)
			}
		}
	}

	slice.Sort(ec2Coll.Instances, func(i, j int) bool {
		return ec2Coll.Instances[i].Name < ec2Coll.Instances[j].Name
	})

	return ec2Coll

}

func (ec *Ec2Collection) Filtered(args []string) []Ec2Instance {

	if len(args) <= 0 {
		return ec.Instances
	}

	var res []Ec2Instance

	for _, v := range ec.Instances {

		match := v.Match(args)

		if match {
			res = append(res, v)
		}

	}

	return res

}
