package cmd

import (
	"fmt"
	"log"

	"github.com/amoydavid/godeployer/internal/config"
	"github.com/amoydavid/godeployer/internal/ssh"
	"github.com/amoydavid/godeployer/internal/tasks"
	"github.com/spf13/cobra"
)

var (
	rollbackConfigFile string
	rollbackVersion    string
)

// rollbackCmd represents the rollback command
var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback to a previous version",
	Long:  `Rollback the application to a previous version on the specified server.`,
	Run:   runRollback,
}

func init() {
	rootCmd.AddCommand(rollbackCmd)
	rollbackCmd.Flags().StringVarP(&rollbackConfigFile, "config", "c", "deploy.yaml", "config file (default is deploy.yaml)")
	rollbackCmd.Flags().StringVarP(&rollbackVersion, "version", "v", "", "version to rollback to (leave empty for immediate previous version)")
}

func runRollback(cmd *cobra.Command, args []string) {
	// 加载配置
	cfg, err := config.LoadConfig(rollbackConfigFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 创建 SSH 客户端
	client, err := ssh.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create SSH client: %v", err)
	}
	defer client.Close()

	// 执行回滚任务
	err = tasks.Rollback(client, cfg, rollbackVersion)
	if err != nil {
		log.Fatalf("Rollback failed: %v", err)
	}

	if rollbackVersion == "" {
		fmt.Println("Rollback to the immediate previous version completed successfully!")
	} else {
		fmt.Printf("Rollback to version %s completed successfully!\n", rollbackVersion)
	}
}
