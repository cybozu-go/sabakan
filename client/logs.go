package client

import (
	"context"
	"io"
	"time"
)

// LogsGet retrieves audit logs.
func LogsGet(ctx context.Context, since, until time.Time, w io.Writer) *Status {
	req := client.NewRequest(ctx, "GET", "logs", nil)
	q := req.URL.Query()
	if !since.IsZero() {
		q.Set("since", since.UTC().Format("20060102"))
	}
	if !until.IsZero() {
		q.Set("until", until.UTC().Format("20060102"))
	}
	req.URL.RawQuery = q.Encode()

	resp, status := client.Do(req)
	if status != nil {
		return status
	}
	defer resp.Body.Close()

	_, err := io.Copy(w, resp.Body)
	if err != nil {
		return ErrorStatus(err)
	}
	return nil
}
