package crossing

import (
	"context"
	"database/sql"
	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
	"errors"
	"time"
)

// RefreshCrossings checks all railroad crossings.
//
//encore:api private method=POST
func RefreshCrossings(ctx context.Context) error {

	url := "http://its.sugarlandtx.gov/api/railmonitor"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	info := retrieveCrossingInfo(ctx, url)
	if info.err != nil {
		return info.err

	}

	rlog.Debug("crossings retrieved", "count", len(info.Resp.Crossings))
	for _, c := range info.Resp.Crossings {
		//isOpen := isCrossingOpen(c.Status)
		//_ = hasCrossingStatusChanged(c.Id, isOpen)
		// if changed publish a notification topic
		if err := AddCrossing(ctx, &AddCrossingParams{Name: c.Name, Latitude: c.Latitude, Longitude: c.Longitude}); err != nil {
			rlog.Error("failed to add crossing", "error", err, "crossing_name", c.Name)
		}
	}

	return nil
}

//encore:api private method=POST
func AddCrossing(ctx context.Context, params *AddCrossingParams) error {
	// check if crossing exists before inserts
	var placeholder int // detect missing row
	err := sqldb.QueryRow(ctx, `
		SELECT 1 FROM crossings
		WHERE name = $1
	`, params.Name).Scan(&placeholder)
	if errors.Is(err, sql.ErrNoRows) {
		if err := insert(ctx, params); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

//encore:api private method=GET
func ListCrossings(ctx context.Context) (*ListCrossing, error) {
	rows, err := sqldb.Query(ctx, `
		SELECT c.id, c.name, c.latitude, c.longitude
		FROM crossings c
		ORDER BY c.id ASC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var crossings []*Crossing
	for rows.Next() {
		var c Crossing
		if err := rows.Scan(&c.Id, &c.Name, &c.Latitude, &c.Longitude); err != nil {
			return nil, err
		}
		crossings = append(crossings, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &ListCrossing{Crossings: crossings}, nil
}

type ListCrossing struct {
	Crossings []*Crossing `json:"crossings"`
}

type AddCrossingParams struct {
	Name      string
	Latitude  float64
	Longitude float64
}

// insert inserts a Crossing into the database.
func insert(ctx context.Context, p *AddCrossingParams) error {
	var _, err = sqldb.Exec(ctx, `
		INSERT INTO crossings (name, latitude, longitude)
		VALUES ($1, $2, $3)
	`, p.Name, p.Latitude, p.Longitude)
	return err
}

type SugarLandCrossings struct {
	UpdateTimestamp string     `json:"update_timestamp"`
	Crossings       []Crossing `json:"crossings"`
}

type Crossing struct {
	Type      string  `json:"type"`
	Id        int     `json:"id"`
	Name      string  `json:"name"`
	Status    string  `json:"status"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

var errCrossingUnresponsive = errors.New("crossing unavailable")
var errCrossingUnauthorized = errors.New("crossing unauthorized/failed")
