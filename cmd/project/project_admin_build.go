package project

import (
	"path/filepath"

	"github.com/haokeyingxiao/haoke-cli/internal/phpexec"
	"github.com/haokeyingxiao/haoke-cli/shop"

	"github.com/haokeyingxiao/haoke-cli/extension"
	"github.com/haokeyingxiao/haoke-cli/logging"
	"github.com/spf13/cobra"
)

var projectAdminBuildCmd = &cobra.Command{
	Use:   "admin-build [project-dir]",
	Short: "Builds the Administration",
	RunE: func(cmd *cobra.Command, args []string) error {
		var projectRoot string
		var err error

		if len(args) == 1 {
			// We need an absolute path for webpack
			projectRoot, err = filepath.Abs(args[0])
			if err != nil {
				return err
			}
		} else if projectRoot, err = findClosestShopwareProject(); err != nil {
			return err
		}

		shopCfg, err := shop.ReadConfig(projectConfigPath, true)
		if err != nil {
			return err
		}

		logging.FromContext(cmd.Context()).Infof("Looking for extensions to build assets in project")

		sources := extension.FindAssetSourcesOfProject(cmd.Context(), projectRoot, shopCfg)

		forceInstall, _ := cmd.PersistentFlags().GetBool("force-install-dependencies")

		shopwareConstraint, err := extension.GetShopwareProjectConstraint(projectRoot)
		if err != nil {
			return err
		}

		assetCfg := extension.AssetBuildConfig{
			DisableStorefrontBuild: true,
			ShopwareRoot:           projectRoot,
			ShopwareVersion:        shopwareConstraint,
			NPMForceInstall:        forceInstall,
			ContributeProject:      extension.IsContributeProject(projectRoot),
		}

		if err := extension.BuildAssetsForExtensions(cmd.Context(), sources, assetCfg); err != nil {
			return err
		}

		skipAssetsInstall, _ := cmd.PersistentFlags().GetBool("skip-assets-install")
		if skipAssetsInstall {
			return nil
		}

		return runTransparentCommand(commandWithRoot(phpexec.ConsoleCommand(cmd.Context(), "assets:install"), projectRoot))
	},
}

func init() {
	projectRootCmd.AddCommand(projectAdminBuildCmd)
	projectAdminBuildCmd.PersistentFlags().Bool("skip-assets-install", false, "Skips the assets installation")
	projectAdminBuildCmd.PersistentFlags().Bool("force-install-dependencies", false, "Force install NPM dependencies")
}
