// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package api provides metal machine management via API.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	metalv1alpha1 "github.com/talos-systems/sidero/app/metal-controller-manager/api/v1alpha1"
)

// Client provides management over simple API.
type Client struct {
	endpoint string
}

// NewClient returns new API client to manage metal machine.
func NewClient(spec metalv1alpha1.ManagementAPI) (*Client, error) {
	return &Client{
		endpoint: spec.Endpoint,
	}, nil
}

func (c *Client) postRequest(path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s%s", c.endpoint, path), nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.Body != nil {
		defer func() {
			_, _ = io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}()
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error: %s", resp.Status)
	}

	return nil
}

// PowerOn will power on a given machine.
func (c *Client) PowerOn() error {
	return c.postRequest("/poweron")
}

// PowerOff will power off a given machine.
func (c *Client) PowerOff() error {
	return c.postRequest("/poweroff")
}

// PowerCycle will power cycle a given machine.
func (c *Client) PowerCycle() error {
	if err := c.postRequest("/poweroff"); err != nil {
		return err
	}

	return c.postRequest("/poweron")
}

// SetPXE makes sure the node will pxe boot next time.
func (c *Client) SetPXE() error {
	return c.postRequest("/pxeboot")
}

// IsPoweredOn checks current power state.
func (c *Client) IsPoweredOn() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://%s/status", c.endpoint), nil)
	if err != nil {
		return false, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}

	if resp.Body != nil {
		defer func() {
			_, _ = io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}()
	}

	var status struct {
		PoweredOn bool
	}

	if err = json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return false, err
	}

	return status.PoweredOn, nil
}
