package instances

import (
	"context"
	"net/url"
	"sort"

	"github.com/henvic/wedeploycli/apihelper"
	"github.com/henvic/wedeploycli/config"
)

// Client for the services
type Client struct {
	*apihelper.Client
}

// New Client
func New(wectx config.Context) *Client {
	return &Client{
		apihelper.New(wectx),
	}
}

// Instance structure
type Instance struct {
	InstanceID string `json:"instanceId"`
	ServiceID  string `json:"serviceId"`
	ProjectID  string `json:"projectId"`
}

// Filter for the instances
type Filter Instance

// Instances structure
type Instances []Instance

// List instances
func (c *Client) List(ctx context.Context, f Filter) (l Instances, err error) {
	var q = url.Values{}

	if f.InstanceID != "" {
		q.Set("instanceId", f.InstanceID)
	}

	if f.ServiceID != "" {
		q.Set("serviceId", f.ServiceID)
	}

	if f.ProjectID != "" {
		q.Set("projectId", f.ProjectID)
	}

	u := "/instances"

	if len(q) != 0 {
		u += "?" + q.Encode()
	}

	err = c.Client.AuthGet(ctx, u, &l)

	sort.Slice(l, func(i, j int) bool {
		return l[i].InstanceID < l[j].InstanceID
	})

	return l, err
}
