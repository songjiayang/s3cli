s3cli
--

`s3cli` is another simple s3 tool. 


## Features

- [x] Multi-environment configuration.
- [x] List buckets.
- [x] List Objects with a bucket.
- [x] Download a object from bucket. 
- [x] Upload a object to a bucket.
- [x] Delete objects from a bucket.    

## Configuration

Default configration path is `./s3cli.json` or `$HOME/s3cli.json`,  the configration content look like:

```json
{
    "default": {
        "region": "cn-east-1",
        "endpoint": "s3-cn-east-1.qiniucs.com",
        "access_key_id": "",
        "secret_key_id": ""
    },
    "aliyun": {
        "region": "oss-cn-hangzhou",
        "endpoint": "oss-cn-hangzhou.aliyuncs.com",
        "access_key_id": "",
        "secret_key_id": ""
    }
}
```

The cli use `default` group for default env, but you can use `--env` to change. 

## Usage

You can use `s3cli --help` or `s3cli SUBCOMMAND --help` to check the usage details.

 ```
 s3cli --help
 
 s3cli is a CLI library for S3 API. 
You can use it to list buckets、list and manage your objects.

Usage:
  s3cli [flags]
  s3cli [command]

Available Commands:
  delete-object delete object with bucket and batch keys, eg: s3cli delete-object BUCKET
  get-object    get object with bucket and key, eg: s3cli get-object BUCKET
  help          Help about any command
  list-bucket   list bucket with keyword, eg: s3cli list-bucket
  list-object   list object with bucket and prefix, eg: s3cli list-object BUCKET
  put-object    put object with bucket and a file, eg: s3cli put-object BUCKET

Flags:
      --config string   config file (default is ./s3cli.json or $HOME/s3cli.json)
      --env string      the env configuration (default "default")
  -h, --help            help for s3cli
      --page int        page to list items (default 1)
      --pagesize int    size of page to list items (default 20)

Use "s3cli [command] --help" for more information about a command.
 ```
 
 ## Supported S3 OSS
 
 - [x] qiniu
 - [x] aliyun
 - [x] aws

s3cli support all oss services that provide s3 protocol.
 
