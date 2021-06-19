package cmd

import (
	"bufio"
	"io"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

func newListObjectCmd() *cobra.Command {
	var (
		extension string
		prefix    string
		maxSize   string
		minSize   string
	)

	var listObjectCmd = &cobra.Command{
		Use:   "list-object",
		Short: "list object with bucket and prefix, eg: s3cli list-object BUCKET",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var (
				bucket string

				startIndex, endIndex       int64 = (page - 1) * pagesize, page*pagesize - 1
				itemIndex                  int64
				data                       [][]string
				minSizeValue, maxSizeValue int64
			)

			bucket = args[0]

			// set min or max size
			if minSize != "" {
				minSizeValue, _ = parseSize(minSize)
			}
			if maxSize != "" {
				maxSizeValue, _ = parseSize(maxSize)
			}

			filterFunc := func(object s3.Object) bool {
				if extension != "" && !strings.HasSuffix(*object.Key, extension) {
					return false
				}

				if minSizeValue > 0 && *object.Size < minSizeValue {
					return false
				}

				if maxSizeValue > 0 && *object.Size > maxSizeValue {
					return false
				}

				return true
			}

			loadObjects := func() {
				var marker string
				for {

					output, err := s3Client.ListObjects(&s3.ListObjectsInput{
						Bucket:  aws.String(bucket),
						Prefix:  aws.String(prefix),
						MaxKeys: aws.Int64(pagesize),
						Marker:  aws.String(marker),
					})
					if err != nil {
						log.Fatal(err)
					}

					for _, object := range output.Contents {
						if !filterFunc(*object) {
							continue
						}

						if itemIndex >= startIndex && itemIndex <= endIndex {
							data = append(data, []string{
								*object.Key,
								humanizeSize(*object.Size),
								object.LastModified.Local().String(),
							})
						}

						itemIndex++

						if itemIndex > endIndex {
							return
						}
					}

					if len(output.Contents) < int(pagesize) {
						return
					}

					marker = *output.Contents[len(output.Contents)-1].Key
				}
			}

			loadObjects()

			printData([]string{"key", "size", "last modified"}, data)
		},
	}

	listObjectCmd.PersistentFlags().StringVar(&prefix, "prefix", "", "list objects with prefix")
	listObjectCmd.PersistentFlags().StringVar(&extension, "ext", "", "list objects with extension")
	listObjectCmd.PersistentFlags().StringVar(&maxSize, "maxsize", "", "list objects with max size, eg: 10MB")
	listObjectCmd.PersistentFlags().StringVar(&minSize, "minsize", "", "list objects with min size, eg: 10KB")

	return listObjectCmd
}

func newGetObjectCmd() *cobra.Command {
	var (
		key    string
		output string
	)

	var getObjectCmd = &cobra.Command{
		Use:   "get-object",
		Short: "get object with bucket and key, eg: s3cli get-object BUCKET",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if key == "" {
				log.Fatal("object key not set")
			}

			if output == "" {
				output = key
			}

			object, err := s3Client.GetObject(&s3.GetObjectInput{
				Bucket: aws.String(args[0]),
				Key:    aws.String(key),
			})
			if err != nil {
				log.Fatal(err)
			}
			defer object.Body.Close()

			f, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

			bar := newBar(int(*object.ContentLength), key, "downloading")

			wl, err := io.Copy(io.MultiWriter(f, bar), object.Body)
			if err != nil {
				log.Fatal(err)
			}

			if wl != *object.ContentLength {
				log.Fatalf("write content length invalid, %d != %d", wl, *object.ContentLength)
			}
		},
	}

	getObjectCmd.PersistentFlags().StringVar(&key, "key", "", "get object with this key")
	getObjectCmd.PersistentFlags().StringVar(&output, "output", "", "the output file of get object")

	return getObjectCmd
}

func newPutObjectCmd() *cobra.Command {
	var (
		key  string
		file string
	)

	var putObjectCmd = &cobra.Command{
		Use:   "put-object",
		Short: "put object with bucket and a file, eg: s3cli put-object BUCKET",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			f, err := os.Open(file)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

			finfo, _ := os.Stat(file)
			bar := newBar(int(finfo.Size()), file, "uploading")
			r := progressbar.NewReader(f, bar)
			_, err = s3Client.PutObject(&s3.PutObjectInput{
				Bucket: aws.String(args[0]),
				Key:    aws.String(key),
				Body: ReaderSeekerBar{
					Reader: &r,
					Seeker: f,
				},
			})

			if err != nil {
				log.Fatal(err)
			}
		},
	}

	putObjectCmd.PersistentFlags().StringVar(&key, "key", "", "the object key to put")
	putObjectCmd.PersistentFlags().StringVar(&file, "file", "", "the file to upload")
	return putObjectCmd
}

type ReaderSeekerBar struct {
	io.Reader
	io.Seeker
}

func newDeleteObjectCmd() *cobra.Command {
	var (
		key     string
		file    string
		confirm bool
	)

	var deleteObjectCmd = &cobra.Command{
		Use:   "delete-object",
		Short: "delete object with bucket and batch keys, eg: s3cli delete-object BUCKET",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			bucket := args[0]
			log.Printf("start delete object in bucket: %s \n", bucket)

			switch {
			case file != "":
				deleteObjectWithFile(bucket, file, confirm)
			case key != "":
				deleteObjectWithKeys(bucket, key, confirm)
			default:
				log.Println("please input object keys to delete")
			}
		},
	}

	deleteObjectCmd.PersistentFlags().StringVar(&key, "key", "", "the object keys to delete")
	deleteObjectCmd.PersistentFlags().StringVar(&file, "file", "", "the object keys stored in a file to delete")
	deleteObjectCmd.PersistentFlags().BoolVar(&confirm, "confirm", false, "delete double confirm")

	return deleteObjectCmd
}

func deleteObjectWithFile(bucket, file string, confirm bool) {
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		objectKey := strings.TrimSpace(scanner.Text())
		doDeleteObject(bucket, objectKey, confirm)
	}
}

func deleteObjectWithKeys(bucket, key string, confirm bool) {
	splits := strings.Split(key, ",")
	for _, objectKey := range splits {
		doDeleteObject(bucket, objectKey, confirm)
	}
}

func doDeleteObject(bucket, objectKey string, confirm bool) {
	log.Printf("delete %s\n", objectKey)

	if !confirm {
		return
	}

	_, err := s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
	})

	if err != nil {
		log.Fatal(err)
	}
}
