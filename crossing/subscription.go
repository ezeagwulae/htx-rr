package crossing

import (
	"context"
	"database/sql"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
	"errors"
)

type SubscriberParams struct {
	CrossingID  int    `json:"crossing_id"`
	PhoneNumber string `json:"phone_number"`
}

//encore:api public method=POST
func Subscribe(ctx context.Context, params *SubscriberParams) error {
	var placeholder int
	err := sqldb.QueryRow(ctx, `
			SELECT 1 FROM crossings
			WHERE id = $1
	`, params.CrossingID).Scan(&placeholder)
	if errors.Is(err, sql.ErrNoRows) {
		rlog.Error("crossing not found", "crossing_id", params.CrossingID)
		return &errs.Error{Code: errs.NotFound, Message: "invalid crossing_id"}
	} else if err != nil {
		return err
	}

	// todo: add db unique constraints on crossing_id and phone_number
	_, err = sqldb.Exec(ctx, `
			INSERT INTO subscriptions (crossing_id, phone_number)
			VALUES ($1, $2)
			`, params.CrossingID, params.PhoneNumber)

	return err
}
