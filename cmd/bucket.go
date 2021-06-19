package cmd

import (
	"log"
	"strings"

	"github.com/spf13/cobra"
)

func newListBucketCmd() *cobra.Command {
	var name string
	var listBucketCmd = &cobra.Command{
		Use:   "list-bucket",
		Short: "list bucket with keyword, eg: s3cli list-bucket",
		Run: func(cmd *cobra.Command, args []string) {
			output, err := s3Client.ListBuckets(nil)
			if err != nil {
				log.Fatal(err)
			}

			var data [][]string

			for _, bucket := range output.Buckets {
				if name == "" || strings.Contains(*bucket.Name, name) {
					data = append(data, []string{*bucket.Name, (*bucket.CreationDate).String()})
				}
			}

			printData([]string{"name", "created date"}, data)
		},
	}

	listBucketCmd.PersistentFlags().StringVar(&name, "name", "", "list bucket with name")

	return listBucketCmd
}
