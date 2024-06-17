package account

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	accountApi "github.com/haokeyingxiao/haoke-cli/account-api"
	"github.com/haokeyingxiao/haoke-cli/logging"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login into your Haoke Account",
	Long:  "",
	RunE: func(cmd *cobra.Command, _ []string) error {
		email := services.Conf.GetAccountEmail()
		password := services.Conf.GetAccountPassword()
		newCredentials := false

		if len(email) == 0 || len(password) == 0 {
			var err error
			email, password, err = askUserForEmailAndPassword()
			if err != nil {
				return err
			}

			newCredentials = true

			if err := services.Conf.SetAccountEmail(email); err != nil {
				return err
			}
			if err := services.Conf.SetAccountPassword(password); err != nil {
				return err
			}
		} else {
			logging.FromContext(cmd.Context()).Infof("Using existing credentials. Use account:logout to logout")
		}

		client, err := accountApi.NewApi(cmd.Context(), accountApi.LoginRequest{Email: email, Password: password})
		if err != nil {
			return fmt.Errorf("login failed with error: %w", err)
		}

		if newCredentials {
			err := services.Conf.Save()
			if err != nil {
				return fmt.Errorf("cannot save config: %w", err)
			}
		}

		profile, err := client.GetMyProfile(cmd.Context())
		if err != nil {
			return err
		}

		logging.FromContext(cmd.Context()).Infof(
			"Hey %s. You are now authenticated and can use all account commands",
			profile.Name,
		)
		return nil
	},
}

func init() {
	accountRootCmd.AddCommand(loginCmd)
}

func askUserForEmailAndPassword() (string, string, error) {
	emailPrompt := promptui.Prompt{
		Label:    "Email",
		Validate: emptyValidator,
	}

	email, err := emailPrompt.Run()
	if err != nil {
		return "", "", fmt.Errorf("prompt failed %w", err)
	}

	passwordPrompt := promptui.Prompt{
		Label:    "Password",
		Validate: emptyValidator,
		Mask:     '*',
	}

	password, err := passwordPrompt.Run()
	if err != nil {
		return "", "", fmt.Errorf("prompt failed %w", err)
	}

	return email, password, nil
}

func emptyValidator(s string) error {
	if len(s) == 0 {
		return fmt.Errorf("this cannot be empty")
	}

	return nil
}
