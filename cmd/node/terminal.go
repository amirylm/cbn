package main

import (
	"fmt"
	"github.com/amirylm/cbn/src/core"
	"github.com/c-bata/go-prompt"
	"io"
	"log"
	"os"
	"strings"
)

func startTerminal(ctrl *core.Controller) error  {
	for {
		t := prompt.Input("> ", completer)

		if strings.TrimSpace(t) == "exit" {
			return nil
		} else {
			t = strings.TrimSuffix(t, "\n")
			fields := strings.Fields(t)
			if len(fields) > 0 {
				action := fields[0]
				fields = fields[1:]

				err := handler(ctrl, action, fields...)
				if err != nil {
					log.Println("Error:", err)
				}
			}
		}
	}
	return nil
}

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "create_bucket <name>", Description: "Create a new bucket"},
		{Text: "bucket_content <hash>", Description: "Get bucket content names"},
		{Text: "buckets", Description: "List buckets"},
		{Text: "upload <bucket> <filepath> <filetype>", Description: "Upload a file"},
		{Text: "download <bucket> <name> <targetpath>", Description: "Download a file"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func handler(ctrl *core.Controller, action string, fields ...string) error {
	switch action {
	case "create_bucket":
		name := fields[0]
		b, err := ctrl.CreateBucket(name, ctrl.Peer().PrivKey())
		if err != nil {
			return err
		}
		raw, err := core.SerializeBucket(b)
		fmt.Println("new bucket was created:", string(raw))
		break
	case "buckets":
		ctrl.ListBuckets(func(b *core.Bucket) bool {
			raw, _ := core.SerializeBucket(b)
			fmt.Println(string(raw))
			return false // don't collect
		})
		break
	case "bucket_content":
		hash := fields[0]
		names, err := ctrl.GetBucketContent(hash)
		if err != nil {
			return err
		}
		fmt.Println(names)
		break
	case "upload":
		bucket := fields[0]
		filepath := fields[1]
		filetype := fields[2]
		f, err := os.Open(filepath)
		if err != nil {
			return err
		}
		stats, _ := f.Stat()
		err = ctrl.Upload(bucket, *core.NewFileHeader(stats.Name(), filetype), f, ctrl.Peer().PrivKey())
		if err != nil {
			return err
		}
		fmt.Println("file was uploaded!")
		break
	case "download":
		bucket := fields[0]
		name := fields[1]
		targetpath := ""
		if len(fields) > 2 {
			targetpath = fields[2]
		}
		reader, _, err := ctrl.Download(bucket, name)
		if err != nil {
			return err
		}
		var writer io.Writer
		if len(targetpath) > 0 {
			f, err := os.Open(targetpath)
			if err != nil {
				return err
			}
			writer = f
		} else {
			writer = os.Stdout
		}
		_, err = io.Copy(writer, reader)
		if err != nil {
			return err
		}
		break
	}
	return nil
}
