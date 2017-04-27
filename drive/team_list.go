package drive

import (
	"fmt"
	"io"
	"text/tabwriter"

	"golang.org/x/net/context"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
)

type TeamListArgs struct {
	Out        io.Writer
	SkipHeader bool
	AbsPath    bool
}

func (self *Drive) TeamList(args TeamListArgs) (err error) {
	listArgs := listAllTeamDrivesArgs{
		fields: []googleapi.Field{"nextPageToken", "teamDrives(id,name)"},
	}
	teamDrives, err := self.listAllTeamDrives(listArgs)
	if err != nil {
		return fmt.Errorf("Failed to list drives: %s", err)
	}

	PrintTeamDriveList(PrintTeamDriveListArgs{
		Out:        args.Out,
		TeamDrives: teamDrives,
		SkipHeader: args.SkipHeader,
	})

	return
}

type listAllTeamDrivesArgs struct {
	fields []googleapi.Field
}

func (self *Drive) listAllTeamDrives(args listAllTeamDrivesArgs) ([]*drive.TeamDrive, error) {
	var teamDrives []*drive.TeamDrive

	var pageSize int64 = 100

	err := self.service.Teamdrives.List().Fields(args.fields...).PageSize(pageSize).Pages(context.TODO(), func(fl *drive.TeamDriveList) error {
		teamDrives = append(teamDrives, fl.TeamDrives...)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return teamDrives, nil
}

type PrintTeamDriveListArgs struct {
	Out        io.Writer
	TeamDrives []*drive.TeamDrive
	SkipHeader bool
}

func PrintTeamDriveList(args PrintTeamDriveListArgs) {
	w := new(tabwriter.Writer)
	w.Init(args.Out, 0, 0, 3, ' ', 0)

	if !args.SkipHeader {
		fmt.Fprintln(w, "Id\tName")
	}

	for _, f := range args.TeamDrives {
		fmt.Fprintf(w, "%s\t%s\n",
			f.Id,
			f.Name)
	}

	w.Flush()
}
