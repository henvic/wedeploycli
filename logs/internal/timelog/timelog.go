package timelog

import 	"time"

// TimeStackDriver format.
// Format:
// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry
// https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#google.protobuf.Timestamp
type TimeStackDriver time.Time

// MarshalJSON is used for writing a JSON value.
func (r TimeStackDriver) MarshalJSON() ([]byte, error) {
	t := time.Time(r)
	return []byte(t.Format(`"` + time.RFC3339Nano + `"`)), nil
}

// UnmarshalJSON is used for parsing a JSON value.
func (r *TimeStackDriver) UnmarshalJSON(data []byte) error {
	// Ignore null, like in the main JSON package.
	if string(data) == "null" {
		return nil
	}

	var rt, err = time.Parse(`"`+time.RFC3339Nano+`"`, string(data))
	*r = TimeStackDriver(rt)
	return err
}
