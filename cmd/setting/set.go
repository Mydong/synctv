package setting

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/synctv-org/synctv/internal/bootstrap"
	"github.com/synctv-org/synctv/internal/settings"
)

var SetCmd = &cobra.Command{
	Use:   "set",
	Short: "set setting",
	Long:  `set setting`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return bootstrap.New(bootstrap.WithContext(cmd.Context())).Add(
			bootstrap.InitDiscardLog,
			bootstrap.InitConfig,
			bootstrap.InitDatabase,
			bootstrap.InitSetting,
		).Run()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return errors.New("args length must be 2")
		}
		s, ok := settings.Settings[args[0]]
		if !ok {
			return errors.New("setting not found")
		}
		err := s.SetString(args[1])
		if err != nil {
			fmt.Printf("set setting %s error: %v\n", args[0], err)
		}
		fmt.Printf("set setting success:\n%s: %v\n", args[0], s.Interface())
		return nil
	},
}

func init() {
	SettingCmd.AddCommand(SetCmd)
}
