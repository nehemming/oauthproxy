/*
Copyright © 2018-2021 Neil Hemming
*/

package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

type (
	secretsFile struct {
		Secrets clientSettings `json:"api"`
	}

	clientSettings struct {
		//TokenUrl request url to the auth source
		TokenURL string `json:"tokenURL,omitempty"`

		// UserName the users name
		UserName string `json:"username,omitempty"`

		// Password password of the user
		Password string `json:"password,omitempty"`

		// ClientID client id associated with the token
		ClientID string `json:"clientid,omitempty"`

		// Client secret
		ClientSecret string `json:"clientsecret,omitempty"`

		//OpenIDScopes specifies optional requested permissions.
		OpenIDScopes []string `json:"scopes,omitempty"`
	}
)

func (cli *cli) requestTokenCmd(cmd *cobra.Command, args []string) error {

	secrets, err := loadSecrets(args[0])
	if err != nil {
		return err
	}

	cfg := oauth2.Config{
		ClientID:     secrets.ClientID,
		ClientSecret: secrets.ClientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: secrets.TokenURL,
			//AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: secrets.OpenIDScopes,
	}

	t, err := cfg.PasswordCredentialsToken(cli.ctx, secrets.UserName, secrets.Password)
	if err != nil {
		return err
	}
	fmt.Println(t.Type(), t.AccessToken)
	return nil
}

func loadSecrets(secretsFilePath string) (*clientSettings, error) {

	b, err := ioutil.ReadFile(secretsFilePath)
	if err != nil {
		return nil, err
	}

	var s secretsFile

	err = json.Unmarshal(b, &s)
	if err != nil {
		return nil, err
	}

	return &s.Secrets, nil
}
