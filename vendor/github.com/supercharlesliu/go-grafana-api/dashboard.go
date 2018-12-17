package gapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type DashboardMeta struct {
	IsStarred bool   `json:"isStarred"`
	Slug      string `json:"slug"`
	Uid       string `json:"uid"`
	Folder    int64  `json:"folderId"`
}

type DashboardSaveResponse struct {
	Slug    string `json:"slug"`
	Id      int64  `json:"id"`
	Uid     string `json:"uid"`
	Status  string `json:"status"`
	Version int64  `json:"version"`
}

const (
	SearchTypeFolder    = "dash-folder"
	SearchTypeDashboard = "dash-db"
)

type SearchResultItem struct {
	Id    int64
	Uid   string
	Title string
	Url   string
	Type  string
	Uri   string
}

func (item *SearchResultItem) IsFolder() bool {
	return item.Type == SearchTypeFolder
}
func (item *SearchResultItem) Slug() string {
	if item.IsFolder() {
		return ""
	}
	return strings.Replace(item.Uri, "db/", "", 1)
}

type Dashboard struct {
	Meta      DashboardMeta          `json:"meta"`
	Model     map[string]interface{} `json:"dashboard"`
	Folder    int64                  `json:"folderId"`
	Overwrite bool                   `json:overwrite`
}

type DashboardData struct {
	Id    int64  `json:"id"`
	Uid   string `json:"uid"`
	Title string `json:"title"`
}
type DashboardVersionItem struct {
	Id      int64 `json:"id"`
	Version int64 `json:"version"`
}
type DashboardVersion struct {
	Id   int64          `json:"id"`
	Data *DashboardData `json:"data"`
}

// Deprecated: use NewDashboard instead
func (c *Client) SaveDashboard(model map[string]interface{}, overwrite bool) (*DashboardSaveResponse, error) {
	wrapper := map[string]interface{}{
		"dashboard": model,
		"overwrite": overwrite,
	}
	data, err := json.Marshal(wrapper)
	if err != nil {
		return nil, err
	}
	req, err := c.newRequest("POST", "/api/dashboards/db", nil, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		data, _ = ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("status: %d, body: %s", resp.StatusCode, data)
	}

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := &DashboardSaveResponse{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func (c *Client) NewDashboard(dashboard Dashboard) (*DashboardSaveResponse, error) {
	data, err := json.Marshal(dashboard)
	if err != nil {
		return nil, err
	}
	req, err := c.newRequest("POST", "/api/dashboards/db", nil, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := &DashboardSaveResponse{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func (c *Client) Dashboard(slug string) (*Dashboard, error) {
	path := fmt.Sprintf("/api/dashboards/db/%s", slug)
	req, err := c.newRequest("GET", path, nil, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := &Dashboard{}
	err = json.Unmarshal(data, &result)
	result.Folder = result.Meta.Folder
	if os.Getenv("GF_LOG") != "" {
		log.Printf("got back dashboard response  %s", data)
	}
	result.Meta.Uid = result.Model["uid"].(string)
	return result, err
}

func (c *Client) DashboardByUid(uid string) (*Dashboard, error) {
	path := fmt.Sprintf("/api/dashboards/uid/%s", uid)
	req, err := c.newRequest("GET", path, nil, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := &Dashboard{}
	err = json.Unmarshal(data, &result)
	result.Folder = result.Meta.Folder
	if os.Getenv("GF_LOG") != "" {
		log.Printf("got back dashboard response  %s", data)
	}
	result.Meta.Uid = result.Model["uid"].(string)
	return result, err
}

func (c *Client) DashboardsByFolder(folderId int64) ([]*SearchResultItem, error) {
	values := url.Values{}
	values.Add("folderIds", strconv.Itoa(int(folderId)))
	values.Add("type", "dash-db")

	req, err := c.newRequest("GET", "/api/search", values, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result []*SearchResultItem
	err = json.Unmarshal(data, &result)
	return result, err
}

func (c *Client) HomeDashboard() (*Dashboard, error) {
	req, err := c.newRequest("GET", "/api/dashboards/home", nil, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := &Dashboard{}
	err = json.Unmarshal(data, &result)
	result.Folder = result.Meta.Folder
	if os.Getenv("GF_LOG") != "" {
		log.Printf("got back dashboard response  %s", data)
	}
	return result, err
}

func (c *Client) DeleteDashboard(slug string) error {
	path := fmt.Sprintf("/api/dashboards/db/%s", slug)
	req, err := c.newRequest("DELETE", path, nil, nil)
	if err != nil {
		return err
	}

	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}

	return nil
}

func (c *Client) GetDashboardVersions(id int64) ([]*DashboardVersionItem, error) {
	var list []*DashboardVersionItem
	path := fmt.Sprintf("/api/dashboards/id/%d/versions/", id)
	req, err := c.newRequest("GET", path, nil, nil)
	if err != nil {
		return list, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return list, err
	}
	if resp.StatusCode != 200 {
		return list, errors.New(resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return list, err
	}

	err = json.Unmarshal(data, &list)
	if err != nil {
		return list, err
	}
	return list, err
}

// Add this api to transform id to uid
func (c *Client) GetDashboardUidById(id int64) (string, error) {
	versions, err := c.GetDashboardVersions(id)
	if err != nil {
		return "", err
	}
	path := fmt.Sprintf("/api/dashboards/id/%d/versions/%d", id, versions[0].Version)
	req, err := c.newRequest("GET", path, nil, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", errors.New(resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result DashboardVersion
	err = json.Unmarshal(data, &result)
	if err != nil {
		return "", err
	}
	return result.Data.Uid, err
}

func (c *Client) GetDashboardIdByUid(uid string) (int64, error) {
	d, err := c.DashboardByUid(uid)
	if err != nil {
		return 0, err
	}
	return int64(d.Model["id"].(float64)), nil
}
