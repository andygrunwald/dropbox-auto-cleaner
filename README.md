# Dropbox auto cleaner

Deletes files from your dropbox folder based on file age.

## Usecase

Several video cameras store their recordings in cloud storages, like Dropbox.
If don't need the recordings after some time, those can be deleted.
This way, you keep your Dropbox storage in shape.

## Usage and configuration

```sh
./dropbox-auto-cleaner -help
Usage of ./dropbox-auto-cleaner:
  -dry
    	If set, the app runs in dry mode. To be deleted files will be printed. No files will be deleted.
  -file-age string
    	File age: Every file inside path that is older than this setting will be deleted. Values from https://pkg.go.dev/time@go1.20.1#ParseDuration are supported. Default: '168h' (aka 7 days). (default "168h")
  -interval string
    	Interval in when the cleaning operation should be triggered. Values from https://pkg.go.dev/time@go1.20.1#ParseDuration are supported. Example value: '24h'. (default "24h")
  -path string
    	Folder path to observe and clean. Example value: '/Apps/Netatmo/Your Name'. Required flag.
```

# Development

Makefile

```sh
build-docker                   Builds the docker image
build                          Compiles the application
help                           Outputs the help
run                            Compiles and starts the application
staticcheck                    Runs static code analyzer staticcheck
vet                            Runs go vet
```