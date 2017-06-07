package drive

import (
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
)

type UploadArgs struct {
	Out         io.Writer
	Progress    io.Writer
	Path        string
	Name        string
	Description string
	Parents     []string
	Folder      string
	Mime        string
	Recursive   bool
	Share       bool
	Delete      bool
	ChunkSize   int64
	Timeout     time.Duration
}

const maxRetries = 5
const timeoutRetryDelay = 30 * time.Second
const errorRetryDelay = 5 * time.Second

func (self *Drive) Upload(args UploadArgs) error {
	if args.ChunkSize > intMax()-1 {
		return fmt.Errorf("Chunk size is to big, max chunk size for this computer is %d", intMax()-1)
	}

	if args.Folder != "" {
		parentId, err := self.parentFromFolder(args.Folder)
		if err != nil {
			return err
		}
		args.Parents = []string{parentId}
	}

	// Ensure that none of the parents are sync dirs
	for _, parent := range args.Parents {
		isSyncDir, err := self.isSyncFile(parent)
		if err != nil {
			return err
		}

		if isSyncDir {
			return fmt.Errorf("%s is a sync directory, use 'sync upload' instead", parent)
		}
	}

	if args.Recursive {
		started := time.Now()
		size, err := self.uploadRecursive(args)
		if err != nil {
			return err
		}
		rate := calcRate(size, started, time.Now())
		log.Printf("Uploaded %s at %s/s\n", formatSize(size, false), formatSize(rate, false))
		return nil
	}

	info, err := os.Stat(args.Path)
	if err != nil {
		return fmt.Errorf("Failed stat file: %s", err)
	}

	if info.IsDir() {
		return fmt.Errorf("'%s' is a directory, use --recursive to upload directories", info.Name())
	}

	f, rate, err := self.uploadFile(args)
	if err != nil {
		return err
	}
	if rate == 0 {
		log.Printf("Skipped %s, already exists\n", f.Id)
	} else {
		log.Printf("Uploaded %s at %s/s, total %s\n", f.Id, formatSize(rate, false), formatSize(f.Size, false))
	}

	if args.Share {
		err = self.shareAnyoneReader(f.Id)
		if err != nil {
			return err
		}

		log.Printf("File is readable by anyone at %s\n", f.WebContentLink)
	}

	if args.Delete {
		err = os.Remove(args.Path)
		if err != nil {
			return fmt.Errorf("Failed to delete file: %s", err)
		}
		log.Printf("Removed %s\n", args.Path)
	}

	return nil
}

func (self *Drive) uploadRecursive(args UploadArgs) (int64, error) {
	var size int64
	info, err := os.Stat(args.Path)
	if err != nil {
		return 0, fmt.Errorf("Failed stat file: %s", err)
	}
	if info.IsDir() {
		args.Name = ""
		size, err = self.uploadDirectory(args)
		if err != nil {
			return 0, err
		}
	} else if info.Mode().IsRegular() {
		f, rate, err := self.uploadFile(args)
		if err != nil {
			return 0, err
		}
		if rate == 0 {
			log.Printf("Skipped %s (%s), already exists\n", args.Path, f.Id)
		} else {
			log.Printf("Uploaded %s at %s/s, total %s\n", f.Id, formatSize(rate, false), formatSize(f.Size, false))
		}
		size = f.Size
	}
	if args.Delete {
		err = os.Remove(args.Path)
		if err != nil {
			return 0, fmt.Errorf("Failed to remove: %s", err)
		}
		log.Printf("Removed %s\n", args.Path)
	}
	return size, nil
}

func (self *Drive) uploadDirectory(args UploadArgs) (int64, error) {
	var totalSize int64
	srcFile, srcFileInfo, err := openFile(args.Path)
	if err != nil {
		return 0, err
	}
	// Close file on function exit
	defer srcFile.Close()

	// check if directory exists
	id, err := self.existingFolderId(args.Parents[0], srcFileInfo.Name())
	if err != nil {
		return 0, err
	}
	if id == "" {
		log.Printf("Creating directory %s\n", srcFileInfo.Name())
		f, err := self.mkdir(MkdirArgs{
			Out:         args.Out,
			Name:        srcFileInfo.Name(),
			Parents:     args.Parents,
			Description: args.Description,
		})
		if err != nil {
			return 0, err
		}
		id = f.Id
	} else {
		log.Printf("Using existing directory %s (%s)\n", srcFileInfo.Name(), id)
	}

	names, err := srcFile.Readdirnames(0)
	if err != nil && err != io.EOF {
		return 0, fmt.Errorf("Failed reading directory: %s", err)
	}

	for _, name := range names {
		newArgs := args
		newArgs.Path = filepath.Join(args.Path, name)
		newArgs.Parents = []string{id}
		newArgs.Description = ""

		size, err := self.uploadRecursive(newArgs)
		if err != nil {
			return 0, err
		}
		totalSize += size
	}

	return totalSize, nil
}

func (self *Drive) uploadFile(args UploadArgs) (*drive.File, int64, error) {
	md5Channel := make(chan string)
	go func() {
		md5Channel <- Md5sum(args.Path)
	}()
	srcFile, srcFileInfo, err := openFile(args.Path)
	if err != nil {
		return nil, 0, err
	}
	defer srcFile.Close()

	dstFile := &drive.File{Description: args.Description}

	if args.Name == "" {
		dstFile.Name = filepath.Base(srcFileInfo.Name())
	} else {
		dstFile.Name = args.Name
	}

	if args.Mime == "" {
		dstFile.MimeType = mime.TypeByExtension(filepath.Ext(dstFile.Name))
	} else {
		dstFile.MimeType = args.Mime
	}

	dstFile.Parents = args.Parents

	// if file exists with same name and checksum, skip upload
	existingFile, err := self.existingFile(dstFile.Parents[0], dstFile.Name)
	if err != nil {
		return nil, 0, err
	}
	if existingFile != nil {
		localMd5 := <-md5Channel
		if localMd5 == existingFile.Md5Checksum {
			return existingFile, 0, nil
		}
	}

	chunkSize := googleapi.ChunkSize(int(args.ChunkSize))
	log.Printf("Uploading %s\n", args.Path)
	started := time.Now()
	var f *drive.File
	retries := 0
	for {
		progressReader := getProgressReader(srcFile, args.Progress, srcFileInfo.Size())
		reader, ctx := getTimeoutReaderContext(progressReader, args.Timeout)
		f, err = self.service.Files.Create(dstFile).SupportsTeamDrives(true).Fields("id", "name", "size", "md5Checksum", "webContentLink").Context(ctx).Media(reader, chunkSize).Do()
		if err != nil {
			if isTimeoutError(err) {
				retries++
				if retries > maxRetries {
					return nil, 0, fmt.Errorf("Failed to upload after %d timeout retries: %s", retries, err)
				}
				log.Printf("Retrying in 30 s after timeout: %s\n", err.Error())
				time.Sleep(timeoutRetryDelay)
			} else if isBackendOrRateLimitError(err) {
				retries++
				if retries > maxRetries {
					return nil, 0, fmt.Errorf("Failed to upload after %d error retries: %s", retries, err)
				}
				log.Printf("Retrying in 5 s after error: %s\n", err.Error())
				time.Sleep(errorRetryDelay)
			} else {
				return nil, 0, fmt.Errorf("Failed to upload file: %s", err)
			}
		} else {
			break
		}
		srcFile.Seek(0, 0)
	}
	localMd5 := <-md5Channel
	if f.Md5Checksum != localMd5 {
		return nil, 0, fmt.Errorf("Failed to verify uploaded file %s from %s, local checksum %s, remote checksum %s", f.Id, args.Path, localMd5, f.Md5Checksum)
	}

	rate := calcRate(f.Size, started, time.Now())

	return f, rate, nil
}

type UploadStreamArgs struct {
	Out         io.Writer
	In          io.Reader
	Name        string
	Description string
	Parents     []string
	Mime        string
	Share       bool
	ChunkSize   int64
	Progress    io.Writer
	Timeout     time.Duration
}

func (self *Drive) UploadStream(args UploadStreamArgs) error {
	if args.ChunkSize > intMax()-1 {
		return fmt.Errorf("Chunk size is to big, max chunk size for this computer is %d", intMax()-1)
	}
	dstFile := &drive.File{Name: args.Name, Description: args.Description}
	if args.Mime != "" {
		dstFile.MimeType = args.Mime
	}
	dstFile.Parents = args.Parents
	chunkSize := googleapi.ChunkSize(int(args.ChunkSize))
	progressReader := getProgressReader(args.In, args.Progress, 0)
	reader, ctx := getTimeoutReaderContext(progressReader, args.Timeout)

	log.Printf("Uploading %s\n", dstFile.Name)
	started := time.Now()

	f, err := self.service.Files.Create(dstFile).SupportsTeamDrives(true).Fields("id", "name", "size", "webContentLink").Context(ctx).Media(reader, chunkSize).Do()
	if err != nil {
		if isTimeoutError(err) {
			return fmt.Errorf("Failed to upload file: timeout, no data was transferred for %v", args.Timeout)
		}
		return fmt.Errorf("Failed to upload file: %s", err)
	}

	rate := calcRate(f.Size, started, time.Now())

	log.Printf("Uploaded %s at %s/s, total %s\n", f.Id, formatSize(rate, false), formatSize(f.Size, false))
	if args.Share {
		err = self.shareAnyoneReader(f.Id)
		if err != nil {
			return err
		}
		log.Printf("File is readable by anyone at %s\n", f.WebContentLink)
	}
	return nil
}

func (self *Drive) parentFromFolder(folderPath string) (string, error) {
	parts := strings.Split(folderPath, "/")
	parentId := ""
	for _, name := range parts {
		if name == "" {
			continue
		}
		if parentId == "" {
			if strings.EqualFold("mydrive", strings.Replace(name, " ", "", -1)) {
				parentId = "root"
			} else {
				result, err := self.service.Teamdrives.List().Fields("teamDrives(id,name)").Do()
				if err != nil {
					return "", err
				}
				for _, drive := range result.TeamDrives {
					if drive.Name == name {
						parentId = drive.Id
						break
					}
				}
			}
			if parentId == "" {
				return "", fmt.Errorf("No top level folder matched name %s", name)
			}
		} else {
			query := fmt.Sprintf("mimeType = 'application/vnd.google-apps.folder' and name = '%s' and '%s' in parents", escapeName(name), parentId)
			result, err := self.service.Files.List().SupportsTeamDrives(true).IncludeTeamDriveItems(true).Q(query).Fields("files(id,name)").Do()
			if err != nil {
				return "", err
			}
			if len(result.Files) == 0 {
				return "", fmt.Errorf("No folders matched name %s", name)
			}
			parentId = result.Files[0].Id
		}
	}
	return parentId, nil
}

func (self *Drive) existingFolderId(parentId string, name string) (string, error) {
	query := fmt.Sprintf("mimeType = 'application/vnd.google-apps.folder' and name = '%s' and '%s' in parents", escapeName(name), parentId)
	file, err := self.fileQuery(query)
	if err != nil {
		return "", err
	}
	if file == nil {
		return "", nil
	}
	return file.Id, nil
}

func (self *Drive) existingFile(parentId string, name string) (*drive.File, error) {
	query := fmt.Sprintf("name = '%s' and '%s' in parents", escapeName(name), parentId)
	return self.fileQuery(query)
}

func (self *Drive) fileQuery(query string) (*drive.File, error) {
	var result *drive.FileList
	retries := 0
	for {
		var err error
		result, err = self.service.Files.List().SupportsTeamDrives(true).IncludeTeamDriveItems(true).Q(query).Fields("files(id,name,md5Checksum)").Do()
		if err != nil {
			if isBackendOrRateLimitError(err) {
				retries++
				if retries > maxRetries {
					return nil, fmt.Errorf("Error finding file: %s", err.Error())
				}
				log.Printf("Retrying in 5 s after find error: %s\n", err.Error())
				time.Sleep(errorRetryDelay)
			} else {
				return nil, err
			}
		} else {
			break
		}
	}
	if len(result.Files) == 0 {
		return nil, nil
	}
	return result.Files[0], nil
}

func escapeName(name string) string {
	return strings.Replace(name, "'", "\\'", -1)
}
