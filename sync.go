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

const (
	DefaultContentType = "application/octet-stream"
)

func syncFiles() {
	a = oss.Auth{config.accessId, config.accessKey}
	o = &oss.OSS{a, config.endpoint, config.endpoint}

	b = o.Bucket(config.bucket)

	remoteFiles, _ := getRemoteFiles()
	localFiles, _ := getLocalFiles()

	for path, file := range localFiles {
		fn := config.prefix + strings.TrimLeft(strings.TrimPrefix(path, config.source), "/")
		if remote, ok := remoteFiles[fn]; ok {
			if file.Mode().IsDir() == false {
				data, err := getFile(path)
				if err == nil {
					hash := hashBytes(data)
					if compareHash(hash, remote.ETag) == false {
						log.Printf("Remote file %s existing, etag %s != %s.\n", fn, remote.ETag, hash)
						putFile(path, fn, data)
					}
				}
			}
		} else {
			if file.Mode().IsDir() == true {
				putDirectory(fn)
			} else {
				data, err := getFile(path)
				if err == nil {
					putFile(path, fn, data)
				}
			}
		}
	}

	if config.deleteIfNotFound {
		for path, _ := range remoteFiles {
			fn := strings.TrimLeft(strings.TrimPrefix(path, config.prefix), "/")
			local := config.source + "/" + fn
			remote := config.prefix + fn
			if _, ok := localFiles[local]; ok == false {
				log.Printf("Remove remote file %s.\n", remote)
				b.Del(remote)
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

func getFile(path string) ([]byte, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("Unable to read local file %s: %s.\n", path, err.Error())
	}
	return data, err
}

func putFile(path string, filename string, data []byte) {
	// get content type
	filetype := mime.TypeByExtension(filepath.Ext(path))
	if filetype == "" {
		log.Printf("Unable to detect content type of file %s, using %s.\n", filename, DefaultContentType)
		filetype = DefaultContentType
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
		get    func(string) error
	)

	files = make(map[string]oss.ContentInfo)

	get = func(prefix string) error {
		remoteFiles, err := b.List(prefix, "/", marker, max)
		if err != nil {
			return err
		}

		for _, p := range remoteFiles.CommonPrefixes {
			marker = ""
			get(p)
		}

		for _, file := range remoteFiles.Contents {
			files[file.Key] = file
		}

		if remoteFiles.IsTruncated {
			marker = remoteFiles.NextMarker
			get(prefix)
		}
		return nil
	}

	err = get(config.prefix)
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
