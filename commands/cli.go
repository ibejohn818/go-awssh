package commands

import (
	"fmt"
	"os"
	"strconv"

	"github.com/ibejohn818/awssh/api"
	"github.com/ibejohn818/awssh/config"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func AddCommands(cmd *cobra.Command, conf *config.AwsshConf) {
	AddLsCmd(cmd, conf)
	AddSshCmd(cmd, conf)
	// AddMusshCmd(cmd, conf)
	AddConfigureCmd(cmd, conf)
	AddIpCmd(cmd, conf)
	AddRunCmd(cmd, conf)
}

func selectInstance(coll []api.Ec2Instance, msg string) api.Ec2Instance {

	for k, v := range coll {

		ln := v.GetLine()
		key := strconv.Itoa((k + 1))

		fmt.Printf("%3v) %v \n", key, ln)

	}

	prompt := promptui.Prompt{
		Label: fmt.Sprintf("%v [%v-%v]", msg, strconv.Itoa(1), strconv.Itoa(len(coll))),
	}

	ans, err := prompt.Run()

	idx, err := strconv.Atoi(ans)

	if err != nil || idx <= 0 || idx > len(coll) {
		fmt.Println("Invalid selection")
		os.Exit(1)
	}

	return coll[(idx - 1)]
}
