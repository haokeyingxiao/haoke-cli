package account

import (
	"fmt"
	accountApi "github.com/haokeyingxiao/haoke-cli/account-api"
	"github.com/haokeyingxiao/haoke-cli/logging"
	"github.com/spf13/cobra"
)

var accountCompanyUseCmd = &cobra.Command{
	Use:   "use [companyId]",
	Short: "Use another company for your Account",
	Args:  cobra.MinimumNArgs(1),
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		companyID := args[0]

		for _, membership := range services.AccountClient.GetMemberships() {
			if membership.Company.Id == companyID {
				if err := services.Conf.SetAccountCompanyId(companyID); err != nil {
					return err
				}

				if err := services.Conf.Save(); err != nil {
					return err
				}

				err := accountApi.InvalidateTokenCache()
				if err != nil {
					return fmt.Errorf("cannot invalidate token cache: %w", err)
				}

				logging.FromContext(cmd.Context()).Infof("Successfully changed your company to %s (%s)", membership.Company.Name, membership.Company.CustomerNumber)
				return nil
			}
		}

		return fmt.Errorf("company with ID \"%s\" not found", companyID)
	},
}

func init() {
	accountCompanyRootCmd.AddCommand(accountCompanyUseCmd)
}
