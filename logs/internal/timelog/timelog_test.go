package timelog

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTimeStackDriverJSONMarshal(t *testing.T) {
	tu := time.Unix(1000000000, 0)

	loc, err := time.LoadLocation("Asia/Shanghai")

	if err != nil {
		panic(err)
	}

	tu = tu.In(loc)

	var past = TimeStackDriver(tu)

	var data []byte
	data, err = json.Marshal(past)

	var want = `"2001-09-09T09:46:40+08:00"`

	if err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	if string(data) != want {
		t.Errorf("Expected time to be %v, got %v instead", want, string(data))
	}
}

func TestTimeStackDriverJSONUnmarshal(t *testing.T) {
	var data = `"2019-06-03T21:31:05.908999919Z"`
	var want = int64(1559597465)
	var r TimeStackDriver

	if err := json.Unmarshal([]byte(data), &r); err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	var rt = time.Time(r)
	var got = rt.Unix()

	if got != want {
		t.Errorf("Expected Unix time %v, got %v instead", want, got)
	}
}

func TestTimeStackDriverJSONUnmarshalNull(t *testing.T) {
	var data = `null`
	var r TimeStackDriver

	if err := json.Unmarshal([]byte(data), &r); err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	var rt = time.Time(r)

	if (rt != time.Time{}) {
		t.Error("Expected null to be ignored")
	}
}
