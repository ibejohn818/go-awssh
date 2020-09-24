package api

import (
	"fmt"
	"os"
	"strconv"

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

			ii = append(ii, *i)
		}
	}

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

	if inst.InstanceID != "" {
		i["InstanceID"] = inst.InstanceID
	}
	if inst.Ip != "" {
		i["Ip"] = inst.Ip
	}
	if inst.Name != "" {
		i["Name"] = inst.Name
	}
	if inst.State != "" {
		i["State"] = inst.State
	}
	return i
}

func (inst *Ec2Instance) GetSubnet() {

}

func (inst *Ec2Instance) GetFormattedLabel() string {
	// tmpl, err := template.New("ListServers").Parse("{{ .HasEip }} [{{ .Ip }}]: {{ .Name }}\n")

	// if err != nil {
	// 	panic(err)
	// }

	// for _, inst := range col.Instances {
	// 	m := inst.GetTplMap()
	// 	tmpl.Execute(os.Stdout, m)
	// }
	t := ""
	return t
}

func SelectInstance(inst []Ec2Instance, msg string) (*Ec2Instance, error) {

	for _, v := range inst {
		fmt.Println(v.Name)
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
