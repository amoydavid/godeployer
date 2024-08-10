package cmd

import (
	"fmt"
	"github.com/amoydavid/godeployer/internal/config"
	"github.com/amoydavid/godeployer/internal/ssh"
	"github.com/amoydavid/godeployer/internal/tasks"
	"github.com/spf13/cobra"
	"log"
)

var configFile string

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy the application",
	Long:  `Deploy the application to the specified server using the configuration provided.`,
	Run:   runDeploy,
}

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.Flags().StringVarP(&configFile, "config", "c", "deploy.yaml", "config file (default is deploy.yaml)")
}

func runDeploy(cmd *cobra.Command, args []string) {
	// 加载配置
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 创建 SSH 客户端
	client, err := ssh.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create SSH client: %v", err)
	}
	defer client.Close()

	// 执行部署任务
	err = tasks.Deploy(client, cfg)
	if err != nil {
		log.Fatalf("Deployment failed: %v", err)
	}

	fmt.Println("Deployment completed successfully!")
}
