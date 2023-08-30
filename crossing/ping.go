package crossing

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type RailRoadCrossingInfo struct {
	Resp       SugarLandCrossings
	LastPollAt time.Time
	Latency    int
	err        error
}

func retrieveCrossingInfo(ctx context.Context) *RailRoadCrossingInfo {
	var (
		url      = "http://its.sugarlandtx.gov/api/railmonitor"
		start    = time.Now()
		resultch = make(chan RailRoadCrossingInfo)
	)

	go func() {
		resp, err := ping(ctx, url)
		defer resp.Body.Close()

		rrInfo := RailRoadCrossingInfo{
			LastPollAt: time.Now(),
			Latency:    int(time.Since(start).Milliseconds()),
		}
		if err != nil {
			rrInfo.err = err
			resultch <- rrInfo
			return
		}
		if resp.StatusCode >= http.StatusBadRequest {
			rrInfo.err = errCrossingUnauthorized
			resultch <- rrInfo
			return
		}

		err = json.NewDecoder(resp.Body).Decode(&rrInfo.Resp)

		resultch <- rrInfo
	}()

	select {
	case <-ctx.Done():
		return &RailRoadCrossingInfo{
			LastPollAt: time.Now(),
			err:        errCrossingUnresponsive,
		}
	case result := <-resultch:
		return &result
	}
}

func ping(ctx context.Context, url string) (*http.Response, error) {
	if !strings.HasPrefix(url, "http:") && !strings.HasPrefix(url, "https:") {
		url = "https://" + url
	}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
