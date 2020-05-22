package common

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	"github.com/spf13/cobra"
)

type ChannelFlags struct {
	Channel string
}

func AddChannelFlags(command *cobra.Command, cFlags *ChannelFlags) {
	command.Flags().StringVar(&cFlags.Channel, "channel", string(shared.GetDefaultChannel()), "Install Kogito operator from Operator hub using provided channel, e.g. (alpha/dev-preview)")
}

func CheckChannelArgs(flags *ChannelFlags) error {
	ch := flags.Channel
	if !shared.IsChannelValid(ch) {
		return fmt.Errorf("Invalid Kogito channel type %s, only alpha/dev-preview channels are allowed ", ch)
	}
	return nil
}
