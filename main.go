package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/plugin"

	"github.com/letscloud-community/packer-plugin-letscloud/builder/letscloud"
	letscloudVersion "github.com/letscloud-community/packer-plugin-letscloud/version"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder(plugin.DEFAULT_NAME, new(letscloud.Builder))
	pps.SetVersion(letscloudVersion.PluginVersion)
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
