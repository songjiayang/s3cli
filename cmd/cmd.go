package cmd

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mitchellh/go-homedir"
	"github.com/olekukonko/tablewriter"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile  string
	env      string
	page     int64
	pagesize int64

	s3Client *s3.S3
	rootCmd  = &cobra.Command{
		Use:   "s3cli",
		Short: "Another simple s3 client tool",
		Long: `s3cli is a CLI library for S3 API. 
You can use it to list buckets、list and manage your objects.`,
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./s3cli.json or $HOME/s3cli.json)")
	rootCmd.PersistentFlags().StringVar(&env, "env", "default", "the env configuration")
	rootCmd.PersistentFlags().Int64VarP(&page, "page", "", 1, "page to list items")
	rootCmd.PersistentFlags().Int64VarP(&pagesize, "pagesize", "", 20, "size of page to list items")

	rootCmd.AddCommand(newListBucketCmd())
	rootCmd.AddCommand(newListObjectCmd())
	rootCmd.AddCommand(newGetObjectCmd())
	rootCmd.AddCommand(newPutObjectCmd())
	rootCmd.AddCommand(newDeleteObjectCmd())
}

func initConfig() {
	viper.SetConfigType("json")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName("s3cli")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Using config file: %s", viper.ConfigFileUsed())
	}

	// init s3client
	s3Conf := &aws.Config{
		Region:           aws.String(viper.GetString(env + ".region")),
		Endpoint:         aws.String(viper.GetString(env + ".endpoint")),
		S3ForcePathStyle: aws.Bool(true),
		Credentials: credentials.NewStaticCredentials(
			viper.GetString(env+".access_key_id"),
			viper.GetString(env+".secret_key_id"), "",
		),
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{Config: *s3Conf}))
	s3Client = s3.New(sess)
}

func printData(headers []string, data [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	for _, v := range data {
		table.Append(v)
	}
	table.Render()
}

const (
	KB = int64(1024)
	MB = KB * KB
	GB = MB * KB
)

var (
	sizeMap = map[string]int64{
		"B":  1,
		"KB": KB,
		"MB": MB,
		"GB": GB,
	}
)

func humanizeSize(size int64) string {
	if size > GB {
		return fmt.Sprintf("%0.2fGB", float64(size/(GB*1.0)))
	}

	if size > MB {
		return fmt.Sprintf("%0.2fMB", float64(size/(MB*1.0)))
	}

	if size > KB {
		return fmt.Sprintf("%0.2fKB", float64(size/(KB*1.0)))
	}

	return fmt.Sprintf("%0.2fB", float64(size))
}

func parseSize(input string) (int64, error) {
	input = strings.ToUpper(input)
	bases := []string{"GB", "MB", "KB", "B"}
	baseValue := int64(1)

	for _, base := range bases {
		if strings.Contains(input, base) {
			input = strings.Replace(input, base, "", -1)
			baseValue = sizeMap[base]
			break
		}
	}

	value, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		return 0, err
	}

	return value * baseValue, nil
}

func newBar(max int, title, tips string) *progressbar.ProgressBar {
	return progressbar.NewOptions(max,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetDescription(fmt.Sprintf("[cyan][%s][reset] %s...", title, tips)),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

}
