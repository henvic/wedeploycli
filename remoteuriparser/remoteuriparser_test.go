package remoteuriparser

import "testing"

type cases struct {
	uri  string
	want string
}

var casesProvider = []cases{
	cases{
		uri:  "address",
		want: "http://api.dashboard.address",
	},
	cases{
		uri:  "address:100",
		want: "http://api.dashboard.address:100",
	},
	cases{
		uri:  "http://address",
		want: "http://api.dashboard.address",
	},
	cases{
		uri:  "http://4.4.4.4",
		want: "http://4.4.4.4",
	},
	cases{
		uri:  "4.4.4.4",
		want: "http://4.4.4.4",
	},
	cases{
		uri:  "4.4.4.4/",
		want: "http://4.4.4.4",
	},
	cases{
		uri:  "4.4.4.4:8000/",
		want: "http://4.4.4.4:8000",
	},
	cases{
		uri:  "wedeploy.io",
		want: "http://api.dashboard.wedeploy.io",
	},
	cases{
		uri:  "https://wedeploy.io",
		want: "https://api.dashboard.wedeploy.io",
	},
	cases{
		uri:  "wedeploy.io:8080/",
		want: "http://api.dashboard.wedeploy.io:8080",
	},
	cases{
		uri:  "wedeploy.io:8080",
		want: "http://api.dashboard.wedeploy.io:8080",
	},
	cases{
		uri:  "http://wedeploy.io:8080",
		want: "http://api.dashboard.wedeploy.io:8080",
	},
	cases{
		uri:  "https://wedeploy.io:8080",
		want: "https://api.dashboard.wedeploy.io:8080",
	},
	cases{
		uri:  "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
		want: "http://2001:0db8:85a3:0000:0000:8a2e:0370:7334",
	},
	cases{
		uri:  "https://2001:0db8:85a3:0000:0000:8a2e:0370:7334",
		want: "https://2001:0db8:85a3:0000:0000:8a2e:0370:7334",
	},
	cases{
		uri:  "foo://2001:0db8:85a3:0000:0000:8a2e:0370:7334",
		want: "foo://2001:0db8:85a3:0000:0000:8a2e:0370:7334",
	},
}

func TestParse(t *testing.T) {
	for _, p := range casesProvider {
		if got := Parse(p.uri); got != p.want {
			t.Errorf("Wanted host %v, got %v instead", p.want, got)
		}
	}
}
