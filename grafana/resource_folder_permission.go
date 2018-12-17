package grafana

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	gapi "github.com/supercharlesliu/go-grafana-api"
)

func ResourceFolderPermission() *schema.Resource {
	return &schema.Resource{
		Create: CreateFolderPermission,
		Read:   ReadFolderPermission,
		Update: UpdateFolderPermission,
		Delete: DeleteFolderPermission,
		Exists: ExistsFolderPermission,
		Importer: &schema.ResourceImporter{
			State: ImportFolderPermission,
		},

		Schema: map[string]*schema.Schema{
			"folder_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"items": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
	}
}

func CreateFolderPermission(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gapi.Client)
	folderId := d.Get("folder_id").(string)
	folderPermission := make([]*gapi.FolderPermission, 0)
	items := d.Get("items").([]interface{})
	for _, v := range items {
		item := v.(map[string]interface{})
		perm := &gapi.FolderPermission{}
		if item["role"] != nil {
			perm.Role = item["role"].(string)
		} else if item["team_id"] != nil {
			teamId, err := strconv.Atoi(item["team_id"].(string))
			if err != nil {
				return err
			}
			perm.TeamId = int64(teamId)
		} else if item["user_id"] != nil {
			userId, err := strconv.Atoi(item["user_id"].(string))
			if err != nil {
				return err
			}
			perm.UserId = int64(userId)
		}

		permId, err := strconv.Atoi(item["permission"].(string))
		if err != nil {
			return err
		}

		permission, err := gapi.NewPermissionType(permId)
		if err != nil {
			return err
		}
		perm.Permission = permission
		folderPermission = append(folderPermission, perm)
	}

	err := client.UpdateFolderPermission(folderId, folderPermission)
	if err != nil {
		return err
	}

	d.SetId(folderId)
	return nil
}

func ReadFolderPermission(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gapi.Client)
	folderId := d.Id()
	folderPermission, err := client.GetFolderPermission(folderId)
	if err != nil {
		return err
	}

	items := make([]map[string]interface{}, 0)
	for _, p := range folderPermission {
		perm := make(map[string]interface{}, 0)
		if p.Role != "" {
			perm["role"] = p.Role
		} else if p.TeamId != 0 {
			perm["team_id"] = strconv.FormatInt(p.TeamId, 10)
		} else if p.UserId != 0 {
			perm["user_id"] = strconv.FormatInt(p.UserId, 10)
		}
		perm["permission"] = p.Permission.String()
		items = append(items, perm)
	}

	d.SetId(folderId)
	d.Set("folder_id", folderId)
	if err := d.Set("items", items); err != nil {
		return err
	}
	return nil
}

func UpdateFolderPermission(d *schema.ResourceData, meta interface{}) error {
	if d.HasChange("items") {
		return CreateFolderPermission(d, meta)
	}
	return nil
}

func DeleteFolderPermission(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gapi.Client)
	folderId := d.Id()
	folderPermission := make([]*gapi.FolderPermission, 0)
	err := client.UpdateFolderPermission(folderId, folderPermission)
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}

func ExistsFolderPermission(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*gapi.Client)
	folderId := d.Id()
	_, err := client.GetFolderPermission(folderId)
	if err != nil && err.Error() == "404 Not Found" {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, err
}

func ImportFolderPermission(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	err := ReadFolderPermission(d, meta)
	if err != nil || d.Id() == "" || d.Get("items") == nil {
		return nil, errors.New(fmt.Sprintf("Error: Unable to import Grafana Dashboard: %s.", err))
	}
	return []*schema.ResourceData{d}, nil
}
