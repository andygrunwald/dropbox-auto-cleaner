package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

const (
	// Timeout we set to abort the process once we run in production mode
	AbortTimeout = 15
)

func main() {
	dropboxPath := flag.String("path", "", "Folder path to observe and clean. Example value: '/Apps/Netatmo/Your Name'. Required flag.")
	tickerInterval := flag.String("interval", "24h", "Interval in when the cleaning operation should be triggered. Values from https://pkg.go.dev/time@go1.20.1#ParseDuration are supported. Example value: '24h'.")
	fileAge := flag.String("file-age", "168h", "File age: Every file inside path that is older than this setting will be deleted. Values from https://pkg.go.dev/time@go1.20.1#ParseDuration are supported. Default: '168h' (aka 7 days).")

	dryRun := flag.Bool("dry", false, "If set, the app runs in dry mode. To be deleted files will be printed. No files will be deleted.")
	flag.Parse()

	dropboxAPIToken := os.Getenv("DROPBOX_CLEANER_API_TOKEN")
	if len(dropboxAPIToken) == 0 {
		log.Fatal("No Dropbox API Token was found. Please ensure the environment variable DROPBOX_CLEANER_API_TOKEN is set correctly.")
	}

	if len(*dropboxPath) == 0 {
		log.Fatal("No Dropbox Folder path was found. Please ensure the environment variable DROPBOX_PATH is set correctly.")
	}

	fileAgeDuration, err := time.ParseDuration(*fileAge)
	if err != nil {
		log.Fatalf("file-age flag value '%s' could not be parsed properly. Only values from https://pkg.go.dev/time@go1.20.1#ParseDuration are supported. Aborting.", *fileAge)
	}

	duration, err := time.ParseDuration(*tickerInterval)
	if err != nil {
		log.Fatalf("interval flag value '%s' could not be parsed properly. Only values from https://pkg.go.dev/time@go1.20.1#ParseDuration are supported. Aborting.", *tickerInterval)
	}

	log.Println("====================")
	log.Println("Dropbox Cleaner")
	log.Println("====================")
	log.Println("Settings under we operate:")
	log.Printf("* Cleaning path: '%s'", *dropboxPath)
	log.Printf("* Every %s", *tickerInterval)
	log.Printf("* Delete files older than %s", *fileAge)
	if *dryRun {
		log.Println("* Not deleting files. Printing them instead. Running in dry run mode")
		log.Println("====================")
	} else {
		log.Println("*****")
		log.Println("* Running in production mode: Files will be deleted")
		log.Printf("* Sleeping for %d seconds to provide you the opportunity to abort the process", AbortTimeout)
		log.Println("* If everything is fine, just wait and the app will do the rest")
		log.Println("====================")

		time.Sleep(AbortTimeout * time.Second)
	}

	dropboxClient := initDropboxFileClient(dropboxAPIToken)

	// Executing first run manually.
	// First tick starts after the configured time has passed.
	firstTick := time.Now().Add(duration).Format(time.RFC1123)
	log.Printf("Starting ticker")
	log.Println("Received interval tick. Executing Dropbox Cleaning operation")
	cleanupDropboxFolder(dropboxClient, *dropboxPath, fileAgeDuration, *dryRun)
	log.Printf("Next interval ticket: %s", firstTick)

	ticker := time.NewTicker(duration)
	quitTicker := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				nextTick := time.Now().Add(duration)
				log.Println("Received interval tick. Executing Dropbox Cleaning operation")

				cleanupDropboxFolder(dropboxClient, *dropboxPath, fileAgeDuration, *dryRun)

				log.Println("Dropbox Cleaning operation done")
				log.Printf("Next interval ticket: %s", nextTick.Format(time.RFC1123))
				// do stuff
			case <-quitTicker:
				log.Println("Signal received to shutdown ticker loop")
				ticker.Stop()
				log.Println("Ticker loop shutdown")
				return
			}
		}
	}()

	// Signal handler to shutdown gracefully.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)
	go func() {
		sig := <-sigs
		log.Printf("Received signal '%s'. Sending signal to shutdown ticker loop", sig)
		close(quitTicker)
		done <- true
	}()

	<-done
	log.Println("Shutting down application")
}

func initDropboxFileClient(dropboxAPIToken string) files.Client {
	config := dropbox.Config{
		Token:    dropboxAPIToken,
		LogLevel: dropbox.LogInfo,
	}
	dropboxClient := files.New(config)

	return dropboxClient
}

func cleanupDropboxFolder(dropboxClient files.Client, dropboxPath string, fileAge time.Duration, dryRun bool) {
	now := time.Now()
	fileAgeTime := now.Add(-fileAge)

	log.Printf("Calling Dropbox API for path '%s'", dropboxPath)
	args := files.NewListFolderArg(dropboxPath)
	args.Recursive = true
	res, err := dropboxClient.ListFolder(args)
	if err != nil {
		log.Printf("Error while calling Dropbox API for path '%s': %+v", dropboxPath, err)
		log.Println("Aborting and skipping this tick.")
		return
	}

	log.Printf("Dropbox API returned %d entries (files and folders) ... Start processing", len(res.Entries))
	for _, v := range res.Entries {
		switch file := v.(type) {
		case *files.FileMetadata:
			// Found a file that is older than the configured duration
			if file.ServerModified.Before(fileAgeTime) {
				age := time.Since(file.ServerModified)
				log.Printf("File %s is %s old", file.Metadata.PathDisplay, age.String())

				if dryRun {
					log.Printf("Dry run enabled: The file '%s' should have been deleted", file.Metadata.PathDisplay)
				} else {
					log.Printf("Deleting file '%s'", file.Metadata.PathDisplay)
					deleteArg := &files.DeleteArg{
						Path: file.Metadata.PathDisplay,
					}
					_, err := dropboxClient.DeleteV2(deleteArg)
					if err != nil {
						log.Printf("Error while deleting '%s': %+v", file.Metadata.PathDisplay, err)
						log.Println("Aborting and skipping this file.")
					}
					log.Printf("Deleting file '%s' ... OK", file.Metadata.PathDisplay)
				}
			}

		case *files.FolderMetadata:
			// This a the folder.
			// Right now, we don't need to take action, when this is a folder.
		}
	}
}
