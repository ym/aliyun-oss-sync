package main

import (
	"flag"
	"github.com/ym/oss-aliyun-go"
	"log"
	"os"
	"strings"
)

type Config struct {
	endpoint         string
	internal         bool
	region           string
	source           string
	bucket           string
	prefix           string
	accessId         string
	accessKey        string
	deleteIfNotFound bool
}

var (
	config Config
)

func getOptions() {
	config = Config{
		endpoint: oss.DefaultEndpoint,
		internal: false,
		region:   "hangzhou",
	}

	flag.StringVar(&config.endpoint, "endpoint", oss.DefaultEndpoint, "oss api endpoint")
	flag.BoolVar(&config.internal, "internal", false, "use internal ip address of oss api")
	flag.StringVar(&config.region, "region", oss.DefaultRegion, "oss api region [hangzhou,qingdao,beijing,hongkong]")
	flag.StringVar(&config.source, "source", "", "source directory")
	flag.StringVar(&config.bucket, "bucket", "", "dest bucket")
	flag.StringVar(&config.prefix, "prefix", "", "dest prefix")
	flag.StringVar(&config.accessId, "key", "", "access key")
	flag.StringVar(&config.accessKey, "secret", "", "access secret")
	flag.BoolVar(&config.deleteIfNotFound, "delete", false, "delete extraneous files from dest bucket")

	flag.Parse()

	if config.endpoint == oss.DefaultEndpoint {
		region := strings.ToLower(config.region)
		switch region {
		case "hangzhou":
			if config.internal {
				config.endpoint = oss.HangzhouInternal
			} else {
				config.endpoint = oss.Hangzhou
			}
		case "qingdao":
			if config.internal {
				config.endpoint = oss.QingdaoInternal
			} else {
				config.endpoint = oss.Qingdao
			}
		case "beijing":
			if config.internal {
				config.endpoint = oss.BeijingInternal
			} else {
				config.endpoint = oss.Beijing
			}
		case "hongkong":
			if config.internal {
				config.endpoint = oss.HongkongInternal
			} else {
				config.endpoint = oss.Hongkong
			}
		default:
			log.Fatalf("Region %s not defined.", config.region)
		}
	}

	if config.accessId == "" || config.accessKey == "" {
		log.Fatalln("Access key and access secret must be specified")
	}

	if config.bucket == "" {
		log.Fatalln("Dest bucket must be specified")
	}

	if strings.HasPrefix(config.prefix, "/") {
		config.prefix = strings.TrimLeft(config.prefix, "/")
	}

	if strings.HasSuffix(config.source, "/") {
		config.source = strings.TrimRight(config.source, "/")
	}

	fi, err := os.Stat(config.source)
	if err != nil {
		log.Fatalf("Unable to stat source directory %s: %s.\n", config.source, err.Error())
	}
	if fi.Mode().IsDir() == false {
		log.Fatalf("Source %s not a directory\n", config.source)
	}
}
