package grafana

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	gapi "github.com/nytm/go-grafana-api"
)

func ResourceUser() *schema.Resource {
	return &schema.Resource{
		Create: CreateUser,
		Read:   ReadUser,
		Update: UpdateUser,
		Delete: DeleteUser,
		Exists: ExistsUser,
		Importer: &schema.ResourceImporter{
			State: ImportUser,
		},

		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"login": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"email": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

// source: https://github.com/gogits/gogs/blob/9ee80e3e5426821f03a4e99fad34418f5c736413/modules/base/tool.go#L58
func GetRandomString(n int, alphabets ...byte) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		if len(alphabets) == 0 {
			bytes[i] = alphanum[b%byte(len(alphanum))]
		} else {
			bytes[i] = alphabets[b%byte(len(alphabets))]
		}
	}
	return string(bytes)
}

func CreateUser(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gapi.Client)

	user := gapi.User{
		Login:   d.Get("login").(string),
		Email:   d.Get("email").(string),
		Name:    d.Get("name").(string),
		IsAdmin: false,
	}

	if user.Password == "" {
		user.Password = GetRandomString(10)
	}

	userId, err := client.CreateUser(user)
	if err != nil {
		return err
	}

	userIdStr := strconv.Itoa(int(userId))
	d.SetId(userIdStr)

	return ReadUser(d, meta)
}

func ExistsUser(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*gapi.Client)
	teamId, _ := strconv.ParseInt(d.Id(), 10, 64)
	_, err := client.User(teamId)
	if err != nil && err.Error() == "404 Not Found" {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, err
}

func ReadUser(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gapi.Client)

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}

	user, err := client.User(id)
	if err != nil {
		if err.Error() == "404 Not Found" {
			log.Printf("[WARN] removing User %d from state because it no longer exists in grafana", id)
			d.SetId("")
			return nil
		}

		return err
	}

	d.SetId(strconv.FormatInt(user.Id, 10))
	d.Set("name", user.Name)
	d.Set("login", user.Login)
	d.Set("email", user.Email)

	return nil
}

func UpdateUser(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gapi.Client)
	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}
	if d.HasChange("name") || d.HasChange("login") || d.HasChange("email") {
		u := gapi.UserUpdate{
			Login: d.Get("login").(string),
			Email: d.Get("email").(string),
			Name:  d.Get("name").(string),
		}
		err := client.UpdateUser(id, u)
		if err != nil {
			return err
		}
	}
	return nil
}

func DeleteUser(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gapi.Client)
	return client.DeleteUser(int64(d.Get("id").(int)))
}

func ImportUser(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	err := ReadUser(d, meta)

	if err != nil || d.Id() == "" {
		return nil, errors.New(fmt.Sprintf("Error: Unable to import Grafana Team: %s.", err))
	}
	return []*schema.ResourceData{d}, nil
}
