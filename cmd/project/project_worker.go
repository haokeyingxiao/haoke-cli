package project

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/haokeyingxiao/haoke-cli/internal/phpexec"
	"github.com/haokeyingxiao/haoke-cli/shop"

	"github.com/spf13/cobra"

	"github.com/haokeyingxiao/haoke-cli/logging"
)

var projectWorkerCmd = &cobra.Command{
	Use:   "worker [amount]",
	Short: "Runs the Symfony Worker in Background",
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		var projectRoot string
		var err error
		workerAmount := 1

		isVerbose, _ := cobraCmd.Flags().GetBool("verbose")
		queuesToConsume, _ := cobraCmd.Flags().GetString("queue")
		memoryLimit, _ := cobraCmd.Flags().GetString("memory-limit")
		timeLimit, _ := cobraCmd.Flags().GetString("time-limit")

		if projectRoot, err = findClosestShopwareProject(); err != nil {
			return err
		}

		if len(args) > 0 {
			workerAmount, err = strconv.Atoi(args[0])

			if err != nil {
				return err
			}
		}

		if memoryLimit == "" {
			memoryLimit = "512M"
		}

		if timeLimit == "" {
			timeLimit = "120"
		}

		cancelCtx, cancel := context.WithCancel(cobraCmd.Context())
		cancelOnTermination(cancelCtx, cancel)

		consumeArgs := []string{"messenger:consume", fmt.Sprintf("--memory-limit=%s", memoryLimit), fmt.Sprintf("--time-limit=%s", timeLimit)}

		if queuesToConsume == "" {
			if is, _ := shop.IsShopwareVersion(projectRoot, ">=6.5.7"); is {
				consumeArgs = append(consumeArgs, "async", "failed", "low_priority")
			} else if is, _ := shop.IsShopwareVersion(projectRoot, ">=6.5"); is {
				consumeArgs = append(consumeArgs, "async", "failed")
			}
		} else {
			consumeArgs = append(consumeArgs, strings.Split(queuesToConsume, ",")...)
		}

		if isVerbose {
			consumeArgs = append(consumeArgs, "-vvv")
		}

		baseName := fmt.Sprintf("shopware-cli-%d", os.Getpid())

		var wg sync.WaitGroup
		for a := 0; a < workerAmount; a++ {
			wg.Add(1)
			go func(ctx context.Context, index int) {
				for {
					cmd := phpexec.ConsoleCommand(cancelCtx, consumeArgs...)
					cmd.Dir = projectRoot
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					cmd.Env = append(os.Environ(), fmt.Sprintf("MESSENGER_CONSUMER_NAME=%s-%d", baseName, index))

					if err := cmd.Run(); err != nil {
						logging.FromContext(ctx).Fatal(err)
					}
				}
			}(cancelCtx, a)
		}

		wg.Wait()

		return nil
	},
}

func init() {
	projectRootCmd.AddCommand(projectWorkerCmd)
	projectWorkerCmd.PersistentFlags().Bool("verbose", false, "Enable verbose output")
	projectWorkerCmd.PersistentFlags().String("queue", "", "Queues to consume")
	projectWorkerCmd.PersistentFlags().String("memory-limit", "", "Memory Limit")
	projectWorkerCmd.PersistentFlags().String("time-limit", "", "Time Limit")
}

func cancelOnTermination(ctx context.Context, cancel context.CancelFunc) {
	logging.FromContext(ctx).Infof("setting up a signal handler")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGTERM)
	go func() {
		logging.FromContext(ctx).Infof("received SIGTERM %v\n", <-s)
		cancel()
	}()
}
