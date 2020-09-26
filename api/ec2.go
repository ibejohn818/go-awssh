package api

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"strconv"

	"github.com/bradfitz/slice"
	"github.com/manifoldco/promptui"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

type ElasticIp struct {
	Ip           *string
	AllocationID *string
}

type Ec2Instance struct {
	InstanceID string
	Name       string
	Ip         string
	PrivateIp  string
	State      string
	HasEip     bool
	AZ         string
}

type Ec2Client struct {
	Sdk ec2iface.EC2API
}

func NewEc2Client(ops ...func(*Ec2Client)) *Ec2Client {
	c := Ec2Client{}

	if len(ops) == 1 {
		ops[0](&c)
	} else {
		c.Sdk = ec2.New(defaultSession(), defaultConfig())
	}

	return &c
}

func (c *Ec2Client) LoadEips() []ElasticIp {

	// sess := session.Must(session.NewSession(&aws.Config{
	// 	Region: aws.String(endpoints.UsWest2RegionID),
	// }))

	inp := &ec2.DescribeAddressesInput{}

	res, err := c.Sdk.DescribeAddresses(inp)

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

func (c *Ec2Client) GetInstances() []Ec2Instance {
	ii := make([]Ec2Instance, 0, 1)

	inp := &ec2.DescribeInstancesInput{}

	res, err := c.Sdk.DescribeInstances(inp)

	if err != nil {
		fmt.Println(err)
	}

	for _, v := range res.Reservations {
		for _, vv := range v.Instances {

			i := &Ec2Instance{}
			if inst := vv.InstanceId; inst != nil {
				i.InstanceID = *inst
			}

			if ip := vv.PublicIpAddress; ip != nil {
				i.Ip = *ip
			}

			if ip := vv.PrivateIpAddress; ip != nil {
				i.PrivateIp = *ip
			}

			if name := parseNameTag(vv.Tags); name != nil {
				i.Name = *name
			}

			if state := vv.State.Name; state != nil {
				i.State = *state
			}
			if az := vv.Placement.AvailabilityZone; az != nil {
				i.AZ = *az
			}

			// check if ip is nil, if just use private
			if len(i.Ip) <= 0 {
				i.Ip = i.PrivateIp
			}

			ii = append(ii, *i)
		}
	}

	slice.Sort(ii, func(i, j int) bool {
		return ii[i].Name < ii[j].Name
	})
	// slice.Sort(ec2Coll.Instances, func(i, j int) bool {
	// 	return ec2Coll.Instances[i].Name < ec2Coll.Instances[j].Name
	// })
	// query, err := inst.Client.
	return ii
}

func parseNameTag(tags []*ec2.Tag) *string {

	var name *string

	for _, tt := range tags {
		if *tt.Key == "Name" {
			name = tt.Value
		}
	}
	return name
}

func (inst *Ec2Instance) GetTplMap() map[string]string {

	var i map[string]string
	i = make(map[string]string)

	i["InstanceID"] = inst.InstanceID
	i["Ip"] = fmt.Sprintf("%15s", inst.Ip)
	i["PrivateIp"] = fmt.Sprintf("%15s", inst.PrivateIp)
	i["Name"] = inst.Name
	i["State"] = inst.State

	return i
}

func (inst *Ec2Instance) GetSubnet() {

}

func (inst *Ec2Instance) GetFormattedLabel(usePrivate bool) string {

	m := inst.GetTplMap()

	m["useIp"] = m["Ip"]

	if usePrivate {
		m["useIp"] = m["PrivateIp"]
	}

	tmpl, err := template.New("ListServers").Parse("[{{ .useIp }}]: {{ .Name }}")

	if err != nil {
		panic(err)
	}

	// for _, inst := range col.Instances {
	// 	m := inst.GetTplMap()
	// 	tmpl.Execute(os.Stdout, m)
	// }

	var buff bytes.Buffer

	tmpl.Execute(&buff, m)

	return buff.String()
}

func SelectInstance(inst []Ec2Instance, msg string, showPrivate bool) (*Ec2Instance, error) {

	for k, v := range inst {
		ln := v.GetFormattedLabel(showPrivate)
		fmt.Printf("%d) %s \n", (k + 1), ln)
	}
	prompt := promptui.Prompt{
		Label: fmt.Sprintf("%s [1-%d]", msg, len(inst)),
		Stdin: os.Stdin,
	}

	ans, _ := prompt.Run()

	key, _ := strconv.Atoi(ans)

	res := inst[(key - 1)]

	return &res, nil

}
