package drive

import (
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/api/drive/v3"
)

const DirectoryMimeType = "application/vnd.google-apps.folder"

type MkdirArgs struct {
	Out         io.Writer
	Name        string
	Description string
	Parents     []string
}

func (self *Drive) Mkdir(args MkdirArgs) error {
	f, err := self.mkdir(args)
	if err != nil {
		return err
	}
	log.Printf("Directory %s created\n", f.Id)
	return nil
}

func (self *Drive) mkdir(args MkdirArgs) (*drive.File, error) {
	dstFile := &drive.File{
		Name:        args.Name,
		Description: args.Description,
		MimeType:    DirectoryMimeType,
	}
	dstFile.Parents = args.Parents
	var f *drive.File
	retries := 0
	const maxRetries = 5
	const errorRetryDelay = 5 * time.Second
	for {
		var err error
		f, err = self.service.Files.Create(dstFile).SupportsTeamDrives(true).Do()
		if err != nil {
			if isBackendOrRateLimitError(err) {
				retries++
				if retries > maxRetries {
					return nil, fmt.Errorf("Failed to create directory after %d error retries: %s", retries, err)
				}
				log.Printf("Retrying create directory in %d s after error: %s\n", errorRetryDelay, err.Error())
				time.Sleep(errorRetryDelay)
			} else {
				return nil, fmt.Errorf("Failed to create directory: %s", err)
			}
		} else {
			break
		}
	}
	return f, nil
}
