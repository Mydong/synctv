package user

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/synctv-org/synctv/internal/bootstrap"
	"github.com/synctv-org/synctv/internal/db"
)

var UnbanCmd = &cobra.Command{
	Use:   "unban",
	Short: "unban user with user id",
	Long:  "unban user with user id",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return bootstrap.New(bootstrap.WithContext(cmd.Context())).Add(
			bootstrap.InitDiscardLog,
			bootstrap.InitConfig,
			bootstrap.InitDatabase,
		).Run()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("missing user id")
		}
		u, err := db.GetUserByID(args[0])
		if err != nil {
			fmt.Printf("get user failed: %s\n", err)
			return nil
		}
		err = db.UnbanUser(u)
		if err != nil {
			fmt.Printf("unban user failed: %s", err)
			return nil
		}
		fmt.Printf("unban user success: %s\n", u.Username)
		return nil
	},
}

func init() {
	UserCmd.AddCommand(UnbanCmd)
}
