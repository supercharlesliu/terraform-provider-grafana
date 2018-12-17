package grafana

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	gapi "github.com/supercharlesliu/go-grafana-api"
)

func ResourceTeam() *schema.Resource {
	return &schema.Resource{
		Create: CreateTeam,
		Read:   ReadTeam,
		Update: UpdateTeam,
		Delete: DeleteTeam,
		Exists: ExistsTeam,
		Importer: &schema.ResourceImporter{
			State: ImportTeam,
		},

		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"members": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
		},
	}
}

func CreateTeam(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gapi.Client)

	model := d.Get("name").(string)

	resp, err := client.NewTeam(model)
	if err != nil {
		return err
	}

	id := strconv.FormatInt(resp.Id, 10)
	d.SetId(id)
	d.Set("name", model)

	// Process members
	set := d.Get("members").(*schema.Set)
	for _, v := range set.List() {
		err = client.AddTeamMember(id, int64(v.(int)))
		if err != nil {
			return err
		}
	}

	return ReadTeam(d, meta)
}

func UpdateTeam(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gapi.Client)
	teamId := d.Id()
	if d.HasChange("name") {
		name := d.Get("name").(string)
		err := client.UpdateTeam(teamId, name)
		if err != nil {
			return err
		}
	}

	memberList, err := getTeamMembers(d, meta)
	if err != nil {
		return err
	}
	memberMap := make(map[int64]bool, 0)
	for _, v := range memberList {
		memberMap[v.(int64)] = true
	}

	if d.HasChange("members") {
		// Add members
		members := d.Get("members").(*schema.Set)
		for _, v := range members.List() {
			id := int64(v.(int))
			if _, found := memberMap[id]; !found {
				err = client.AddTeamMember(d.Id(), id)
				if err != nil {
					return err
				}
			}
			delete(memberMap, id)
		}

		// Remove members
		for userId, _ := range memberMap {
			err = client.RemoveTeamMember(d.Id(), userId)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func getTeamMembers(d *schema.ResourceData, meta interface{}) ([]interface{}, error) {
	client := meta.(*gapi.Client)
	members, err := client.TeamMembers(d.Id())
	if err != nil {
		return nil, err
	}
	memberList := make([]interface{}, 0)
	for _, member := range members {
		memberList = append(memberList, member.UserId)
	}
	return memberList, nil
}

func ExistsTeam(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*gapi.Client)
	teamId, _ := strconv.ParseInt(d.Id(), 10, 64)
	_, err := client.Team(teamId)
	if err != nil && err.Error() == "404 Not Found" {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, err
}

func ReadTeam(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gapi.Client)

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}

	team, err := client.Team(id)
	if err != nil {
		if err.Error() == "404 Not Found" {
			log.Printf("[WARN] removing Team %d from state because it no longer exists in grafana", id)
			d.SetId("")
			return nil
		}

		return err
	}

	d.SetId(strconv.FormatInt(team.Id, 10))
	d.Set("name", team.Name)

	// deal with team members
	memberList, err := getTeamMembers(d, meta)
	set := schema.NewSet(func(v interface{}) int {
		return hashcode.String(strconv.Itoa(int(v.(int64))))
	}, memberList)
	d.Set("members", set)
	return nil
}

func DeleteTeam(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gapi.Client)
	return client.DeleteTeam(d.Id())
}

func ImportTeam(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	err := ReadTeam(d, meta)

	if err != nil || d.Id() == "" {
		return nil, errors.New(fmt.Sprintf("Error: Unable to import Grafana Team: %s.", err))
	}
	return []*schema.ResourceData{d}, nil
}
