package servicenow

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type Records struct {
	Result []map[string]interface{} `json:"result"`
}

var (
	linkRE *regexp.Regexp
)

func init() {
	linkRE = regexp.MustCompile(`<([^>]+)>\s*;\s*rel\s*=\s*"([^"]+)"`)
}

func (snp *ServiceNowProvider) getPaginatedResults(
	client *http.Client,
	url string,
	result interface{},
) (string, error) {
	snp.Interop.Logger.Debugf(
		"making servicenow request using URL %s...",
		url,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	if snp.AuthType == AUTH_TYPE_BASIC {
		req.SetBasicAuth(snp.ApiUser, snp.ApiPassword)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("fetch results failed: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	snp.Interop.Logger.Debugf(
		"read %d bytes, unmarshaling JSON...",
		len(body),
	)

	err = json.Unmarshal(body, result)
	if err != nil {
		return "", err
	}

	linkHeader := resp.Header.Get("Link")
	if linkHeader != "" {
		all := linkRE.FindAllStringSubmatch(linkHeader, -1)
		for _, tag := range all {
			if tag[2] == "next" {
				return tag[1], nil
			}
		}
	}

	return "", nil
}

func (snp *ServiceNowProvider) getRecords(
	tableName string,
	query string,
	extraQueryParms string,
	fields []string,
) (
	[]map[string]interface{},
	error,
) {
	var results []map[string]interface{}

	client, err := snp.createHttpClient()
	if err != nil {
		return nil, err
	}

	done := false

	sysparmQuery := ""
	if query != "" {
		sysparmQuery = "&sysparm_query=" + url.QueryEscape(query)
	}

	url := fmt.Sprintf(
		"%s/api/now/table/%s?sysparm_fields=sys_id,%s&sysparm_limit=%d&sysparm_offset=0%s",
		snp.ApiURL,
		tableName,
		strings.Join(fields, ","),
		snp.PageSize,
		sysparmQuery,
	)

	if extraQueryParms != "" {
		url = url+ extraQueryParms
	}

	for !done {
		records := &Records{}

		nextUrl, err := snp.getPaginatedResults(client, url, records)
		if err != nil {
			return nil, err
		}

		results = append(results, records.Result...)

		if nextUrl == "" {
			done = true
			continue
		}

		url = nextUrl
	}
	return results, nil
}

func (snp *ServiceNowProvider) createHttpClient() (*http.Client, error) {
	if snp.AuthType == AUTH_TYPE_OAUTH {
		ctx := context.TODO()

		if snp.OAuthGrantType == OAUTH_GRANT_TYPE_PASSWORD {
			endpointParams := url.Values{}
			endpointParams.Add("username", snp.ApiUser)
			endpointParams.Add("password", snp.ApiPassword)

			oauthConfig := &oauth2.Config{
				ClientID:     snp.OAuthClientID,
				ClientSecret: snp.OAuthClientSecret,
				Endpoint: oauth2.Endpoint{
					AuthURL:   "",
					TokenURL:  snp.OAuthTokenURL,
					AuthStyle: oauth2.AuthStyleAutoDetect,
				},
				Scopes: snp.OAuthScopes,
			}

			token, err := oauthConfig.PasswordCredentialsToken(
				ctx,
				snp.ApiUser,
				snp.ApiPassword,
			)
			if err != nil {
				return nil, err
			}

			return oauthConfig.Client(ctx, token), nil
		}

		oauthConfig := &clientcredentials.Config{
			ClientID:     snp.OAuthClientID,
			ClientSecret: snp.OAuthClientSecret,
			TokenURL:     snp.OAuthTokenURL,
			Scopes:       snp.OAuthScopes,
		}

		return oauthConfig.Client(ctx), nil
	}

	return &http.Client{}, nil
}
