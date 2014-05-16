package main

import (
	"github.com/ym/oss-aliyun-go"
	"io/ioutil"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

var (
	a   oss.Auth
	o   *oss.OSS
	b   *oss.Bucket
	err error
)

func syncFiles() {
	a = oss.Auth{config.accessId, config.accessKey}
	o = &oss.OSS{a, config.endpoint, config.endpoint}

	b = o.Bucket(config.bucket)

	remoteFiles, _ := getRemoteFiles()
	localFiles, _ := getLocalFiles()

	for path, file := range localFiles {
		fn := config.prefix + strings.TrimLeft(strings.TrimSuffix(path, config.source), "/")
		if remote, ok := remoteFiles[fn]; ok {
			log.Printf("Remote file %s existing, calculating local etag %s.\n", fn, remote.ETag)
		} else {
			if file.Mode().IsDir() == true {
				putDirectory(fn)
			} else {
				putFile(path, fn)
			}
		}
	}
}

func putDirectory(filename string) {
	data := []byte("")
	err := b.Put(filename, data, "text/plain", oss.Private)
	if err != nil {
		log.Printf("Unable to create directory %s: %s.\n", filename, err.Error())
	} else {
		log.Printf("Created directory %s.\n", filename)
	}
}

func putFile(path string, filename string) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("Unable to read local file %s: %s.\n", path, err.Error())
	}
	filetype := mime.TypeByExtension(filepath.Ext(path))
	if filetype == "" {
		filetype = "application/octet-stream"
	}
	err = b.Put(filename, data, filetype, oss.Private)
	if err != nil {
		log.Printf("Unable to upload file %s: %s.\n", filename, err.Error())
	} else {
		log.Printf("Uploaded file: %s ... ", filename)
	}
}

func getRemoteFiles() (files map[string]oss.ContentInfo, err error) {
	var (
		marker = ""
		max    = 512
		get    func() error
	)

	files = make(map[string]oss.ContentInfo)

	get = func() error {
		remoteFiles, err := b.List(config.prefix, "/", marker, max)
		if err != nil {
			return err
		}
		for _, file := range remoteFiles.Contents {
			files[file.Key] = file
		}
		if remoteFiles.IsTruncated {
			marker = remoteFiles.NextMarker
			get()
		}
		return nil
	}

	err = get()
	if err != nil {
		return files, err
	}

	return files, nil
}

func getLocalFiles() (files map[string]os.FileInfo, err error) {
	files = make(map[string]os.FileInfo)
	err = filepath.Walk(config.source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("%s while looking up %s", err.Error(), path)
			return err
		}
		files[path] = info
		return nil
	})
	return files, err
}
