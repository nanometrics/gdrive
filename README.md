godrive
======


## Overview
godrive is a command line utility for interacting with Google Drive.

## Prerequisites
None, binaries are statically linked.
If you want to compile from source you need the [go toolchain](http://golang.org/doc/install).
Version 1.5 or higher.

## Installation
### With [Homebrew](http://brew.sh) on Mac
```
brew install godrive
```
### Other
Download `godrive` from one of the links below. On unix systems
run `chmod +x godrive` after download to make the binary executable.
The first time godrive is launched (i.e. run `godrive about` in your
terminal not just `godrive`), you will be prompted for a verification code.
The code is obtained by following the printed url and authenticating with the
google account for the drive you want access to. This will create a token file
inside the .godrive folder in your home directory. Note that anyone with access
to this file will also have access to your google drive.
If you want to manage multiple drives you can use the global `--config` flag
or set the environment variable `GODRIVE_CONFIG_DIR`.
Example: `GODRIVE_CONFIG_DIR="/home/user/.godrive-secondary" godrive list`
You will be prompted for a new verification code if the folder does not exist.

## Compile from source
```bash
go get github.com/nanometrics/godrive
```
The godrive binary should now be available at `$GOPATH/bin/godrive`


## Godrive 2
Godrive 2 is more or less a full rewrite and is not backwards compatible
with godrive 1 as all the command line arguments has changed slightly.
Godrive 2 uses version 3 of the google drive api and my google-api-go-client
fork is no longer needed.

### Syncing
Godrive 2 supports basic syncing. It only syncs one way at the time and works
more like rsync than e.g. dropbox. Files that are synced to google drive
are tagged with an appProperty so that the files on drive can be traversed
faster. This means that you can't upload files with `godrive upload` into
a sync directory as the files would be missing the sync tag, and would be
ignored by the sync commands.
The current implementation is slow and uses a lot of memory if you are
syncing many files. Currently only one file is uploaded at the time,
the speed can be improved in the future by uploading several files concurrently.
To learn more see usage and the examples below.

### Service Account
For server to server communication, where user interaction is not a viable option,
is it possible to use a service account, as described in this [Google document](https://developers.google.com/identity/protocols/OAuth2ServiceAccount).
If you want to use a service account, instead of being interactively prompted for
authentication, you need to use the `--service-account <serviceAccountCredentials>`
global option, where `serviceAccountCredentials` is a file in JSON format obtained
through the Google API Console, and its location is relative to the config dir.

#### .godriveignore
Placing a .godriveignore in the root of your sync directory can be used to
skip certain files from being synced. .godriveignore follows the same
rules as [.gitignore](https://git-scm.com/docs/gitignore), except that godrive only reads the .godriveignore file in the root of the sync directory, not ones in any subdirectories.


## Usage
```
godrive [global] list [options]                                 List files
godrive [global] download [options] <fileId>                    Download file or directory
godrive [global] download query [options] <query>               Download all files and directories matching query
godrive [global] upload [options] <path>                        Upload file or directory
godrive [global] upload - [options] <name>                      Upload file from stdin
godrive [global] update [options] <fileId> <path>               Update file, this creates a new revision of the file
godrive [global] info [options] <fileId>                        Show file info
godrive [global] mkdir [options] <name>                         Create directory
godrive [global] share [options] <fileId>                       Share file or directory
godrive [global] share list <fileId>                            List files permissions
godrive [global] share revoke <fileId> <permissionId>           Revoke permission
godrive [global] delete [options] <fileId>                      Delete file or directory
godrive [global] sync list [options]                            List all syncable directories on drive
godrive [global] sync content [options] <fileId>                List content of syncable directory
godrive [global] sync download [options] <fileId> <path>        Sync drive directory to local directory
godrive [global] sync upload [options] <path> <fileId>          Sync local directory to drive
godrive [global] changes [options]                              List file changes
godrive [global] revision list [options] <fileId>               List file revisions
godrive [global] revision download [options] <fileId> <revId>   Download revision
godrive [global] revision delete <fileId> <revId>               Delete file revision
godrive [global] import [options] <path>                        Upload and convert file to a google document, see 'about import' for available conversions
godrive [global] export [options] <fileId>                      Export a google document
godrive [global] about [options]                                Google drive metadata, quota usage
godrive [global] about import                                   Show supported import formats
godrive [global] about export                                   Show supported export formats
godrive version                                                 Print application version
godrive help                                                    Print help
godrive help <command>                                          Print command help
godrive help <command> <subcommand>                             Print subcommand help
```

#### List files
```
godrive [global] list [options]

global:
  -c, --config <configDir>         Application path, default: /Users/<user>/.godrive
  --refresh-token <refreshToken>   Oauth refresh token used to get access token (for advanced users)
  --access-token <accessToken>     Oauth access token, only recommended for short-lived requests because of short lifetime (for advanced users)
  --service-account <accountFile>  Oauth service account filename, used for server to server communication without user interaction (file is relative to config dir)

options:
  -m, --max <maxFiles>       Max files to list, default: 30
  -q, --query <query>        Default query: "trashed = false and 'me' in owners". See https://developers.google.com/drive/search-parameters
  --order <sortOrder>        Sort order. See https://godoc.org/google.golang.org/api/drive/v3#FilesListCall.OrderBy
  --name-width <nameWidth>   Width of name column, default: 40, minimum: 9, use 0 for full width
  --absolute                 Show absolute path to file (will only show path from first parent)
  --no-header                Dont print the header
  --bytes                    Size in bytes
```

List file in subdirectory


```
./godrive list --query " 'IdOfTheParentFolder' in parents"
```

#### Download file or directory
```
godrive [global] download [options] <fileId>

global:
  -c, --config <configDir>         Application path, default: /Users/<user>/.godrive
  --refresh-token <refreshToken>   Oauth refresh token used to get access token (for advanced users)
  --access-token <accessToken>     Oauth access token, only recommended for short-lived requests because of short lifetime (for advanced users)
  --service-account <accountFile>  Oauth service account filename, used for server to server communication without user interaction (file is relative to config dir)

options:
  -f, --force           Overwrite existing file
  -r, --recursive       Download directory recursively, documents will be skipped
  --path <path>         Download path
  --delete              Delete remote file when download is successful
  --no-progress         Hide progress
  --stdout              Write file content to stdout
  --timeout <timeout>   Set timeout in seconds, use 0 for no timeout. Timeout is reached when no data is transferred in set amount of seconds, default: 300
```

#### Download all files and directories matching query
```
godrive [global] download query [options] <query>

global:
  -c, --config <configDir>         Application path, default: /Users/<user>/.godrive
  --refresh-token <refreshToken>   Oauth refresh token used to get access token (for advanced users)
  --access-token <accessToken>     Oauth access token, only recommended for short-lived requests because of short lifetime (for advanced users)
  --service-account <accountFile>  Oauth service account filename, used for server to server communication without user interaction (file is relative to config dir)

options:
  -f, --force       Overwrite existing file
  -r, --recursive   Download directories recursively, documents will be skipped
  --path <path>     Download path
  --no-progress     Hide progress
```

#### Upload file or directory
```
godrive [global] upload [options] <path>

global:
  -c, --config <configDir>         Application path, default: /Users/<user>/.godrive
  --refresh-token <refreshToken>   Oauth refresh token used to get access token (for advanced users)
  --access-token <accessToken>     Oauth access token, only recommended for short-lived requests because of short lifetime (for advanced users)
  --service-account <accountFile>  Oauth service account filename, used for server to server communication without user interaction (file is relative to config dir)

options:
  -r, --recursive               Upload directory recursively
  -p, --parent <parent>         Parent id, used to upload file to a specific directory, can be specified multiple times to give many parents
  --name <name>                 Filename
  --description <description>   File description
  --no-progress                 Hide progress
  --mime <mime>                 Force mime type
  --share                       Share file
  --delete                      Delete local file when upload is successful
  --timeout <timeout>           Set timeout in seconds, use 0 for no timeout. Timeout is reached when no data is transferred in set amount of seconds, default: 300
  --chunksize <chunksize>       Set chunk size in bytes, default: 8388608
```

#### Upload file from stdin
```
godrive [global] upload - [options] <name>

global:
  -c, --config <configDir>         Application path, default: /Users/<user>/.godrive
  --refresh-token <refreshToken>   Oauth refresh token used to get access token (for advanced users)
  --access-token <accessToken>     Oauth access token, only recommended for short-lived requests because of short lifetime (for advanced users)
  --service-account <accountFile>  Oauth service account filename, used for server to server communication without user interaction (file is relative to config dir)

options:
  -p, --parent <parent>         Parent id, used to upload file to a specific directory, can be specified multiple times to give many parents
  --chunksize <chunksize>       Set chunk size in bytes, default: 8388608
  --description <description>   File description
  --mime <mime>                 Force mime type
  --share                       Share file
  --timeout <timeout>           Set timeout in seconds, use 0 for no timeout. Timeout is reached when no data is transferred in set amount of seconds, default: 300
  --no-progress                 Hide progress
```

#### Update file, this creates a new revision of the file
```
godrive [global] update [options] <fileId> <path>

global:
  -c, --config <configDir>         Application path, default: /Users/<user>/.godrive
  --refresh-token <refreshToken>   Oauth refresh token used to get access token (for advanced users)
  --access-token <accessToken>     Oauth access token, only recommended for short-lived requests because of short lifetime (for advanced users)
  --service-account <accountFile>  Oauth service account filename, used for server to server communication without user interaction (file is relative to config dir)

options:
  -p, --parent <parent>         Parent id, used to upload file to a specific directory, can be specified multiple times to give many parents
  --name <name>                 Filename
  --description <description>   File description
  --no-progress                 Hide progress
  --mime <mime>                 Force mime type
  --timeout <timeout>           Set timeout in seconds, use 0 for no timeout. Timeout is reached when no data is transferred in set amount of seconds, default: 300
  --chunksize <chunksize>       Set chunk size in bytes, default: 8388608
```

#### Show file info
```
godrive [global] info [options] <fileId>

global:
  -c, --config <configDir>         Application path, default: /Users/<user>/.godrive
  --refresh-token <refreshToken>   Oauth refresh token used to get access token (for advanced users)
  --access-token <accessToken>     Oauth access token, only recommended for short-lived requests because of short lifetime (for advanced users)
  --service-account <accountFile>  Oauth service account filename, used for server to server communication without user interaction (file is relative to config dir)

options:
  --bytes   Show size in bytes
```

#### Create directory
```
godrive [global] mkdir [options] <name>

global:
  -c, --config <configDir>         Application path, default: /Users/<user>/.godrive
  --refresh-token <refreshToken>   Oauth refresh token used to get access token (for advanced users)
  --access-token <accessToken>     Oauth access token, only recommended for short-lived requests because of short lifetime (for advanced users)
  --service-account <accountFile>  Oauth service account filename, used for server to server communication without user interaction (file is relative to config dir)

options:
  -p, --parent <parent>         Parent id of created directory, can be specified multiple times to give many parents
  --description <description>   Directory description
```

#### Share file or directory
```
godrive [global] share [options] <fileId>

global:
  -c, --config <configDir>         Application path, default: /Users/<user>/.godrive
  --refresh-token <refreshToken>   Oauth refresh token used to get access token (for advanced users)
  --access-token <accessToken>     Oauth access token, only recommended for short-lived requests because of short lifetime (for advanced users)
  --service-account <accountFile>  Oauth service account filename, used for server to server communication without user interaction (file is relative to config dir)

options:
  --role <role>     Share role: owner/writer/commenter/reader, default: reader
  --type <type>     Share type: user/group/domain/anyone, default: anyone
  --email <email>   The email address of the user or group to share the file with. Requires 'user' or 'group' as type
  --discoverable    Make file discoverable by search engines
  --revoke          Delete all sharing permissions (owner roles will be skipped)
```

#### List files permissions
```
godrive [global] share list <fileId>

global:
  -c, --config <configDir>         Application path, default: /Users/<user>/.godrive
  --refresh-token <refreshToken>   Oauth refresh token used to get access token (for advanced users)
  --access-token <accessToken>     Oauth access token, only recommended for short-lived requests because of short lifetime (for advanced users)
  --service-account <accountFile>  Oauth service account filename, used for server to server communication without user interaction (file is relative to config dir)
```

#### Revoke permission
```
godrive [global] share revoke <fileId> <permissionId>

global:
  -c, --config <configDir>         Application path, default: /Users/<user>/.godrive
  --refresh-token <refreshToken>   Oauth refresh token used to get access token (for advanced users)
  --access-token <accessToken>     Oauth access token, only recommended for short-lived requests because of short lifetime (for advanced users)
  --service-account <accountFile>  Oauth service account filename, used for server to server communication without user interaction (file is relative to config dir)
```

#### Delete file or directory
```
godrive [global] delete [options] <fileId>

global:
  -c, --config <configDir>         Application path, default: /Users/<user>/.godrive
  --refresh-token <refreshToken>   Oauth refresh token used to get access token (for advanced users)
  --access-token <accessToken>     Oauth access token, only recommended for short-lived requests because of short lifetime (for advanced users)
  --service-account <accountFile>  Oauth service account filename, used for server to server communication without user interaction (file is relative to config dir)

options:
  -r, --recursive   Delete directory and all it's content
```

#### List all syncable directories on drive
```
godrive [global] sync list [options]

global:
  -c, --config <configDir>         Application path, default: /Users/<user>/.godrive
  --refresh-token <refreshToken>   Oauth refresh token used to get access token (for advanced users)
  --access-token <accessToken>     Oauth access token, only recommended for short-lived requests because of short lifetime (for advanced users)
  --service-account <accountFile>  Oauth service account filename, used for server to server communication without user interaction (file is relative to config dir)

options:
  --no-header   Dont print the header
```

#### List content of syncable directory
```
godrive [global] sync content [options] <fileId>

global:
  -c, --config <configDir>         Application path, default: /Users/<user>/.godrive
  --refresh-token <refreshToken>   Oauth refresh token used to get access token (for advanced users)
  --access-token <accessToken>     Oauth access token, only recommended for short-lived requests because of short lifetime (for advanced users)
  --service-account <accountFile>  Oauth service account filename, used for server to server communication without user interaction (file is relative to config dir)

options:
  --order <sortOrder>        Sort order. See https://godoc.org/google.golang.org/api/drive/v3#FilesListCall.OrderBy
  --path-width <pathWidth>   Width of path column, default: 60, minimum: 9, use 0 for full width
  --no-header                Dont print the header
  --bytes                    Size in bytes
```

#### Sync drive directory to local directory
```
godrive [global] sync download [options] <fileId> <path>

global:
  -c, --config <configDir>         Application path, default: /Users/<user>/.godrive
  --refresh-token <refreshToken>   Oauth refresh token used to get access token (for advanced users)
  --access-token <accessToken>     Oauth access token, only recommended for short-lived requests because of short lifetime (for advanced users)
  --service-account <accountFile>  Oauth service account filename, used for server to server communication without user interaction (file is relative to config dir)

options:
  --keep-remote         Keep remote file when a conflict is encountered
  --keep-local          Keep local file when a conflict is encountered
  --keep-largest        Keep largest file when a conflict is encountered
  --delete-extraneous   Delete extraneous local files
  --dry-run             Show what would have been transferred
  --no-progress         Hide progress
  --timeout <timeout>   Set timeout in seconds, use 0 for no timeout. Timeout is reached when no data is transferred in set amount of seconds, default: 300
```

#### Sync local directory to drive
```
godrive [global] sync upload [options] <path> <fileId>

global:
  -c, --config <configDir>         Application path, default: /Users/<user>/.godrive
  --refresh-token <refreshToken>   Oauth refresh token used to get access token (for advanced users)
  --access-token <accessToken>     Oauth access token, only recommended for short-lived requests because of short lifetime (for advanced users)
  --service-account <accountFile>  Oauth service account filename, used for server to server communication without user interaction (file is relative to config dir)

options:
  --keep-remote             Keep remote file when a conflict is encountered
  --keep-local              Keep local file when a conflict is encountered
  --keep-largest            Keep largest file when a conflict is encountered
  --delete-extraneous       Delete extraneous remote files
  --dry-run                 Show what would have been transferred
  --no-progress             Hide progress
  --timeout <timeout>       Set timeout in seconds, use 0 for no timeout. Timeout is reached when no data is transferred in set amount of seconds, default: 300
  --chunksize <chunksize>   Set chunk size in bytes, default: 8388608
```

#### List file changes
```
godrive [global] changes [options]

global:
  -c, --config <configDir>         Application path, default: /Users/<user>/.godrive
  --refresh-token <refreshToken>   Oauth refresh token used to get access token (for advanced users)
  --access-token <accessToken>     Oauth access token, only recommended for short-lived requests because of short lifetime (for advanced users)
  --service-account <accountFile>  Oauth service account filename, used for server to server communication without user interaction (file is relative to config dir)

options:
  -m, --max <maxChanges>     Max changes to list, default: 100
  --since <pageToken>        Page token to start listing changes from
  --now                      Get latest page token
  --name-width <nameWidth>   Width of name column, default: 40, minimum: 9, use 0 for full width
  --no-header                Dont print the header
```

#### List file revisions
```
godrive [global] revision list [options] <fileId>

global:
  -c, --config <configDir>         Application path, default: /Users/<user>/.godrive
  --refresh-token <refreshToken>   Oauth refresh token used to get access token (for advanced users)
  --access-token <accessToken>     Oauth access token, only recommended for short-lived requests because of short lifetime (for advanced users)
  --service-account <accountFile>  Oauth service account filename, used for server to server communication without user interaction (file is relative to config dir)

options:
  --name-width <nameWidth>   Width of name column, default: 40, minimum: 9, use 0 for full width
  --no-header                Dont print the header
  --bytes                    Size in bytes
```

#### Download revision
```
godrive [global] revision download [options] <fileId> <revId>

global:
  -c, --config <configDir>         Application path, default: /Users/<user>/.godrive
  --refresh-token <refreshToken>   Oauth refresh token used to get access token (for advanced users)
  --access-token <accessToken>     Oauth access token, only recommended for short-lived requests because of short lifetime (for advanced users)
  --service-account <accountFile>  Oauth service account filename, used for server to server communication without user interaction (file is relative to config dir)

options:
  -f, --force           Overwrite existing file
  --no-progress         Hide progress
  --stdout              Write file content to stdout
  --path <path>         Download path
  --timeout <timeout>   Set timeout in seconds, use 0 for no timeout. Timeout is reached when no data is transferred in set amount of seconds, default: 300
```

#### Delete file revision
```
godrive [global] revision delete <fileId> <revId>

global:
  -c, --config <configDir>         Application path, default: /Users/<user>/.godrive
  --refresh-token <refreshToken>   Oauth refresh token used to get access token (for advanced users)
  --access-token <accessToken>     Oauth access token, only recommended for short-lived requests because of short lifetime (for advanced users)
  --service-account <accountFile>  Oauth service account filename, used for server to server communication without user interaction (file is relative to config dir)
```

#### Upload and convert file to a google document, see 'about import' for available conversions
```
godrive [global] import [options] <path>

global:
  -c, --config <configDir>         Application path, default: /Users/<user>/.godrive
  --refresh-token <refreshToken>   Oauth refresh token used to get access token (for advanced users)
  --access-token <accessToken>     Oauth access token, only recommended for short-lived requests because of short lifetime (for advanced users)
  --service-account <accountFile>  Oauth service account filename, used for server to server communication without user interaction (file is relative to config dir)

options:
  -p, --parent <parent>   Parent id, used to upload file to a specific directory, can be specified multiple times to give many parents
  --no-progress           Hide progress
```

#### Export a google document
```
godrive [global] export [options] <fileId>

global:
  -c, --config <configDir>         Application path, default: /Users/<user>/.godrive
  --refresh-token <refreshToken>   Oauth refresh token used to get access token (for advanced users)
  --access-token <accessToken>     Oauth access token, only recommended for short-lived requests because of short lifetime (for advanced users)
  --service-account <accountFile>  Oauth service account filename, used for server to server communication without user interaction (file is relative to config dir)

options:
  -f, --force     Overwrite existing file
  --mime <mime>   Mime type of exported file
  --print-mimes   Print available mime types for given file
```

#### Google drive metadata, quota usage
```
godrive [global] about [options]

global:
  -c, --config <configDir>         Application path, default: /Users/<user>/.godrive
  --refresh-token <refreshToken>   Oauth refresh token used to get access token (for advanced users)
  --access-token <accessToken>     Oauth access token, only recommended for short-lived requests because of short lifetime (for advanced users)
  --service-account <accountFile>  Oauth service account filename, used for server to server communication without user interaction (file is relative to config dir)

options:
  --bytes   Show size in bytes
```

#### Show supported import formats
```
godrive [global] about import

global:
  -c, --config <configDir>         Application path, default: /Users/<user>/.godrive
  --refresh-token <refreshToken>   Oauth refresh token used to get access token (for advanced users)
  --access-token <accessToken>     Oauth access token, only recommended for short-lived requests because of short lifetime (for advanced users)
  --service-account <accountFile>  Oauth service account filename, used for server to server communication without user interaction (file is relative to config dir)
```

#### Show supported export formats
```
godrive [global] about export

global:
  -c, --config <configDir>         Application path, default: /Users/<user>/.godrive
  --refresh-token <refreshToken>   Oauth refresh token used to get access token (for advanced users)
  --access-token <accessToken>     Oauth access token, only recommended for short-lived requests because of short lifetime (for advanced users)
  --service-account <accountFile>  Oauth service account filename, used for server to server communication without user interaction (file is relative to config dir)
```


## Examples
#### List files
```
$ godrive list
Id                             Name                    Type   Size     Created
0B3X9GlR6EmbnZ3gyeGw4d3ozbUk   drive-windows-x64.exe   bin    6.6 MB   2015-07-18 16:43:58
0B3X9GlR6EmbnTXlSc1FqV1dvSTQ   drive-windows-386.exe   bin    5.2 MB   2015-07-18 16:43:53
0B3X9GlR6EmbnVjIzMDRqck1aekE   drive-osx-x64           bin    6.5 MB   2015-07-18 16:43:50
0B3X9GlR6EmbnbEpXdlhza25zT1U   drive-osx-386           bin    5.2 MB   2015-07-18 16:43:41
0B3X9GlR6Embnb095MGxEYmJhY2c   drive-linux-x64         bin    6.5 MB   2015-07-18 16:43:38
```

#### List largest files
```
$ godrive list --query "name contains 'godrive'" --order "quotaBytesUsed desc" -m 3
Id                             Name                     Type   Size     Created
0B3X9GlR6EmbnZXpDRG1xblM2LTg   godrive-linux-mips64      bin    8.5 MB   2016-02-22 21:07:04
0B3X9GlR6EmbnNW5CTV8xdFkxTjg   godrive-linux-mips64le    bin    8.5 MB   2016-02-22 21:07:07
0B3X9GlR6EmbnZ1NGS25FdEVlWEk   godrive-osx-x64           bin    8.3 MB   2016-02-21 20:22:13
```

#### Upload file
```
$ godrive upload godrive-osx-x64
Uploading godrive-osx-x64
Uploaded 0B3X9GlR6EmbnZ1NGS25FdEVlWEk at 3.8 MB/s, total 8.3 MB
```

#### Make directory
```
$ godrive mkdir godrive-bin
Directory 0B3X9GlR6EmbnY1RLVTk5VUtOVkk created
```

#### Upload file to directory
```
$ godrive upload --parent 0B3X9GlR6EmbnY1RLVTk5VUtOVkk godrive-osx-x64
Uploading godrive-osx-x64
Uploaded 0B3X9GlR6EmbnNTk0SkV0bm5Hd0E at 2.5 MB/s, total 8.3 MB
```

#### Download file
```
$ godrive download 0B3X9GlR6EmbnZ1NGS25FdEVlWEk
Downloading godrive-osx-x64 -> godrive-osx-x64
Downloaded 0B3X9GlR6EmbnZ1NGS25FdEVlWEk at 8.3 MB/s, total 8.3 MB
```

#### Share a file
```
$ godrive share 0B3X9GlR6EmbnNTk0SkV0bm5Hd0E
Granted reader permission to anyone
```

#### Pipe content directly to google drive
```
$ echo "Hello World" | godrive upload - hello.txt
Uploading hello.txt
Uploaded 0B3X9GlR6EmbnaXVrOUpIcWlUS0E at 8.0 B/s, total 12.0 B
```

#### Print file to stdout
```
$ godrive download --stdout 0B3X9GlR6EmbnaXVrOUpIcWlUS0E
Hello World
```

#### Get file info
```
$ godrive info 0B3X9GlR6EmbnNTk0SkV0bm5Hd0E
Id: 0B3X9GlR6EmbnNTk0SkV0bm5Hd0E
Name: godrive-osx-x64
Path: godrive-bin/godrive-osx-x64
Mime: application/octet-stream
Size: 8.3 MB
Created: 2016-02-21 20:47:04
Modified: 2016-02-21 20:47:04
Md5sum: b607f29231a3b2d16098c4212516470f
Shared: True
Parents: 0B3X9GlR6EmbnY1RLVTk5VUtOVkk
ViewUrl: https://drive.google.com/file/d/0B3X9GlR6EmbnNTk0SkV0bm5Hd0E/view?usp=drivesdk
DownloadUrl: https://docs.google.com/uc?id=0B3X9GlR6EmbnNTk0SkV0bm5Hd0E&export=download
```

#### Update file (create new revision)
```
$ godrive update 0B3X9GlR6EmbnNTk0SkV0bm5Hd0E godrive-osx-x64
Uploading godrive-osx-x64
Updated 0B3X9GlR6EmbnNTk0SkV0bm5Hd0E at 2.0 MB/s, total 8.3 MB
```

#### List file revisions
```
$ godrive revision list 0B3X9GlR6EmbnNTk0SkV0bm5Hd0E
Id                                                    Name             Size     Modified              KeepForever
0B3X9GlR6EmbnOFlHSTZQNWJWMGN2ckZucC9VaEUwczV1cUNrPQ   godrive-osx-x64   8.3 MB   2016-02-21 20:47:04   False
0B3X9GlR6EmbndVEwMlZCUldGWUlPb2lTS25rOFo1L2t6c2ZVPQ   godrive-osx-x64   8.3 MB   2016-02-21 21:12:09   False
```

#### Download revision
```
$ godrive revision download 0B3X9GlR6EmbnNTk0SkV0bm5Hd0E 0B3X9GlR6EmbnOFlHSTZQNWJWMGN2ckZucC9VaEUwczV1cUNrPQ
Downloading godrive-osx-x64 -> godrive-osx-x64
Download complete, rate: 8.3 MB/s, total size: 8.3 MB
```

#### Export google doc as docx
```
$ godrive export --mime application/vnd.openxmlformats-officedocument.wordprocessingml.document 1Kt5A8X7X2RQrEi5t6Y9W1LayRc4hyrFiG63y2dIJEvk
Exported 'foo.docx' with mime type: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document'
```

#### Import csv as google spreadsheet
```
$ godrive import foo.csv
Imported 1mTl3DjIvap4tpTX_oMkDcbDT8ShtiGJRlozTfkXpeko with mime type: 'application/vnd.google-apps.spreadsheet'
```

#### Syncing directory to drive
```
# Create directory on drive
$ godrive mkdir drive-bin
Directory 0B3X9GlR6EmbnOEd6cEh6bU9XZWM created

# Sync to drive
$ godrive sync upload _release/bin 0B3X9GlR6EmbnOEd6cEh6bU9XZWM
Starting sync...
Collecting local and remote file information...
Found 32 local files and 0 remote files

6 remote directories are missing
[0001/0006] Creating directory drive-bin/bsd
[0002/0006] Creating directory drive-bin/linux
[0003/0006] Creating directory drive-bin/osx
[0004/0006] Creating directory drive-bin/plan9
[0005/0006] Creating directory drive-bin/solaris
[0006/0006] Creating directory drive-bin/windows

26 remote files are missing
[0001/0026] Uploading bsd/godrive-dragonfly-x64 -> drive-bin/bsd/godrive-dragonfly-x64
[0002/0026] Uploading bsd/godrive-freebsd-386 -> drive-bin/bsd/godrive-freebsd-386
[0003/0026] Uploading bsd/godrive-freebsd-arm -> drive-bin/bsd/godrive-freebsd-arm
[0004/0026] Uploading bsd/godrive-freebsd-x64 -> drive-bin/bsd/godrive-freebsd-x64
[0005/0026] Uploading bsd/godrive-netbsd-386 -> drive-bin/bsd/godrive-netbsd-386
[0006/0026] Uploading bsd/godrive-netbsd-arm -> drive-bin/bsd/godrive-netbsd-arm
[0007/0026] Uploading bsd/godrive-netbsd-x64 -> drive-bin/bsd/godrive-netbsd-x64
[0008/0026] Uploading bsd/godrive-openbsd-386 -> drive-bin/bsd/godrive-openbsd-386
[0009/0026] Uploading bsd/godrive-openbsd-arm -> drive-bin/bsd/godrive-openbsd-arm
[0010/0026] Uploading bsd/godrive-openbsd-x64 -> drive-bin/bsd/godrive-openbsd-x64
[0011/0026] Uploading linux/godrive-linux-386 -> drive-bin/linux/godrive-linux-386
[0012/0026] Uploading linux/godrive-linux-arm -> drive-bin/linux/godrive-linux-arm
[0013/0026] Uploading linux/godrive-linux-arm64 -> drive-bin/linux/godrive-linux-arm64
[0014/0026] Uploading linux/godrive-linux-mips64 -> drive-bin/linux/godrive-linux-mips64
[0015/0026] Uploading linux/godrive-linux-mips64le -> drive-bin/linux/godrive-linux-mips64le
[0016/0026] Uploading linux/godrive-linux-ppc64 -> drive-bin/linux/godrive-linux-ppc64
[0017/0026] Uploading linux/godrive-linux-ppc64le -> drive-bin/linux/godrive-linux-ppc64le
[0018/0026] Uploading linux/godrive-linux-x64 -> drive-bin/linux/godrive-linux-x64
[0019/0026] Uploading osx/godrive-osx-386 -> drive-bin/osx/godrive-osx-386
[0020/0026] Uploading osx/godrive-osx-arm -> drive-bin/osx/godrive-osx-arm
[0021/0026] Uploading osx/godrive-osx-x64 -> drive-bin/osx/godrive-osx-x64
[0022/0026] Uploading plan9/godrive-plan9-386 -> drive-bin/plan9/godrive-plan9-386
[0023/0026] Uploading plan9/godrive-plan9-x64 -> drive-bin/plan9/godrive-plan9-x64
[0024/0026] Uploading solaris/godrive-solaris-x64 -> drive-bin/solaris/godrive-solaris-x64
[0025/0026] Uploading windows/godrive-windows-386.exe -> drive-bin/windows/godrive-windows-386.exe
[0026/0026] Uploading windows/godrive-windows-x64.exe -> drive-bin/windows/godrive-windows-x64.exe
Sync finished in 1m18.891946279s

# Add new local file
$ echo "google drive binaries" > _release/bin/readme.txt

# Sync again
$ godrive sync upload _release/bin 0B3X9GlR6EmbnOEd6cEh6bU9XZWM
Starting sync...
Collecting local and remote file information...
Found 33 local files and 32 remote files

1 remote files are missing
[0001/0001] Uploading readme.txt -> drive-bin/readme.txt
Sync finished in 2.201339535s

# Modify local file
$ echo "for all platforms" >> _release/bin/readme.txt

# Sync again
$ godrive sync upload _release/bin 0B3X9GlR6EmbnOEd6cEh6bU9XZWM
Starting sync...
Collecting local and remote file information...
Found 33 local files and 33 remote files

1 local files has changed
[0001/0001] Updating readme.txt -> drive-bin/readme.txt
Sync finished in 1.890244258s
```

#### List content of sync directory
```
$ godrive sync content 0B3X9GlR6EmbnOEd6cEh6bU9XZWM
Id                             Path                             Type   Size     Modified
0B3X9GlR6EmbnMldxMFV1UGVMTlE   bsd                              dir             2016-02-21 22:54:00
0B3X9GlR6EmbnM05sQ3hVUnJnOXc   bsd/godrive-dragonfly-x64         bin    7.8 MB   2016-02-21 22:54:14
0B3X9GlR6EmbnVy1KXzA4dlU5RVE   bsd/godrive-freebsd-386           bin    6.1 MB   2016-02-21 22:54:18
0B3X9GlR6Embnb29QQkFtSlRiZnc   bsd/godrive-freebsd-arm           bin    6.1 MB   2016-02-21 22:54:20
0B3X9GlR6EmbnMkFQYVpSaHhHTXM   bsd/godrive-freebsd-x64           bin    7.8 MB   2016-02-21 22:54:23
0B3X9GlR6EmbnVmJRMl9hUDloVU0   bsd/godrive-netbsd-386            bin    6.1 MB   2016-02-21 22:54:25
0B3X9GlR6EmbnLVlTZWpxOEF4Q2s   bsd/godrive-netbsd-arm            bin    6.1 MB   2016-02-21 22:54:28
0B3X9GlR6EmbnOENUZmh3anJmNG8   bsd/godrive-netbsd-x64            bin    7.8 MB   2016-02-21 22:54:30
0B3X9GlR6EmbnWTRoQ2ZVQXRfQlU   bsd/godrive-openbsd-386           bin    6.1 MB   2016-02-21 22:54:32
0B3X9GlR6EmbncEtlN3ZuQ0VUWms   bsd/godrive-openbsd-arm           bin    6.1 MB   2016-02-21 22:54:35
0B3X9GlR6EmbnMlFLY1ptNEFyZWc   bsd/godrive-openbsd-x64           bin    7.8 MB   2016-02-21 22:54:38
0B3X9GlR6EmbncGtSajQyNzloVEE   linux                            dir             2016-02-21 22:54:01
0B3X9GlR6EmbnMWVudkJmb1NZdmM   linux/godrive-linux-386           bin    6.1 MB   2016-02-21 22:54:40
0B3X9GlR6Embnbnpla1R2VHV5T2M   linux/godrive-linux-arm           bin    6.1 MB   2016-02-21 22:54:42
0B3X9GlR6EmbnM0s2cU1YWkNJSjA   linux/godrive-linux-arm64         bin    7.7 MB   2016-02-21 22:54:45
0B3X9GlR6EmbnNU9NNi1TdDc4S2c   linux/godrive-linux-mips64        bin    8.5 MB   2016-02-21 22:54:47
0B3X9GlR6EmbnSmdQNjRKZ2dWV1U   linux/godrive-linux-mips64le      bin    8.5 MB   2016-02-21 22:54:50
0B3X9GlR6EmbnS0g0OVgxMHY5Z3c   linux/godrive-linux-ppc64         bin    7.8 MB   2016-02-21 22:54:52
0B3X9GlR6EmbneVp6ZXRpR3FhWlU   linux/godrive-linux-ppc64le       bin    7.8 MB   2016-02-21 22:54:54
0B3X9GlR6EmbnczdJT195dFVxdU0   linux/godrive-linux-x64           bin    7.8 MB   2016-02-21 22:54:57
0B3X9GlR6EmbnTXZXeDRnSDdVS1E   osx                              dir             2016-02-21 22:54:02
0B3X9GlR6EmbnWnRheXJNR0pUMU0   osx/godrive-osx-386               bin    6.6 MB   2016-02-21 22:54:59
0B3X9GlR6EmbnRzNqMWFXdDR1Rms   osx/godrive-osx-arm               bin    6.6 MB   2016-02-21 22:55:01
0B3X9GlR6EmbnaDlVWTZDd0JIeEU   osx/godrive-osx-x64               bin    8.3 MB   2016-02-21 22:55:04
0B3X9GlR6EmbnWW84UFBvbHlURXM   plan9                            dir             2016-02-21 22:54:02
0B3X9GlR6EmbnTmc0a2RNdDZDRUU   plan9/godrive-plan9-386           bin    5.8 MB   2016-02-21 22:55:07
0B3X9GlR6EmbnT1pYZ2p4Sk9FVFk   plan9/godrive-plan9-x64           bin    7.4 MB   2016-02-21 22:55:10
0B3X9GlR6EmbnbnZnXzlYVHoxdk0   readme.txt                       bin    40.0 B   2016-02-21 22:59:56
0B3X9GlR6EmbnSWF1QUlta3RnaGc   solaris                          dir             2016-02-21 22:54:03
0B3X9GlR6EmbnaWFOV0YxSGs5Znc   solaris/godrive-solaris-x64       bin    7.7 MB   2016-02-21 22:55:13
0B3X9GlR6EmbnNE5ySkEzbWQ4Qms   windows                          dir             2016-02-21 22:54:03
0B3X9GlR6EmbnX1RIT2w1TWZYWFU   windows/godrive-windows-386.exe   bin    6.1 MB   2016-02-21 22:55:15
0B3X9GlR6EmbndmVMU05POGRPS3c   windows/godrive-windows-x64.exe   bin    7.8 MB   2016-02-21 22:55:18
```
