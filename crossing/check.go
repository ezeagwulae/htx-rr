package crossing

import (
	"context"
	"encore.dev/cron"
	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
	"errors"
	"golang.org/x/sync/errgroup"
	"strings"
	"time"
)

type UpdateCheckParams struct {
	Name   string
	Status string
}

//encore:api private method=PUT
func UpdateCheck(ctx context.Context, p *UpdateCheckParams) error {
	//todo: implement update
	_, err := sqldb.Exec(ctx, `
				UPDATE checks  
				SET status = $1, checked_at = $2
				WHERE crossing_id = (SELECT id from crossings WHERE name = $3)
	`, p.Status, time.Now(), p.Name)
	if err != nil {
		return err
	}
	return nil
}

// Refresh all tracked crossings every 6 hours.
// This is unlikely to change much, can later be changed to 24 hours
var _ = cron.NewJob("refresh-crossings", cron.JobConfig{
	Title:    "Refresh all railroad crossing",
	Endpoint: RefreshCrossings,
	Every:    6 * cron.Hour,
})

var _ = cron.NewJob("check-crossing-status", cron.JobConfig{
	Title:    "Checks the status of railroad crossing",
	Endpoint: CheckRailroadStatus,
	Every:    5 * cron.Minute,
})

//encore:api private method=POST
func CheckRailroadStatus(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	info := retrieveCrossingInfo(ctx)
	if info.err != nil {
		return info.err

	}

	g, ctx := errgroup.WithContext(context.Background())
	g.SetLimit(10) // concurrently check up to 10 crossings
	for _, cx := range info.Resp.Crossings {
		cx := cx
		g.Go(func() error {
			return checkStatus(ctx, &cx)
		})
	}
	return g.Wait()
}

func checkStatus(ctx context.Context, c *Crossing) error {
	prevStatus, err := getPreviousStatus(ctx, c.Name)
	if err != nil {
		return err
	}

	// do nothing if status is unchanged
	if isCrossingOpen(prevStatus) == isCrossingOpen(c.Status) {
		return nil
	}

	// todo: publish notification topic
	rlog.Debug("crossing status changed", "crossing", c.Name, "prev", prevStatus, "curr", c.Status)
	if err := UpdateCheck(ctx, &UpdateCheckParams{Name: c.Name, Status: c.Status}); err != nil {
		return err
	}
	return nil
}

// getPreviousStatus reports if the previous known crossing status was open or closed
// based on the last known state
func getPreviousStatus(ctx context.Context, crossingName string) (string, error) {
	var status string
	err := sqldb.QueryRow(ctx, `
			SELECT ck.status
			FROM checks ck
			INNER JOIN crossings cx on cx.id = ck.crossing_id
			WHERE cx.name = $1
	`, crossingName).Scan(&status)
	if errors.Is(err, sqldb.ErrNoRows) {
		// There was no previous status; treat this as if the crossing was open before
		return "", nil
	} else if err != nil {
		return "", err
	}

	return status, nil
}

func isCrossingOpen(status string) bool {
	openStatuses := []string{
		"clear", "active", "operational",
		"offline", // means stop is not functional
	}
	if stringInSlice(status, openStatuses) {
		return true
	}
	return false
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == strings.ToLower(a) {
			return true
		}
	}
	return false
}
