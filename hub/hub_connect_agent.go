package hub

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
)

// GenerateConnectManifest asks the gkehub API for a gke-connect-agent manifest
func (c *Client) GenerateConnectManifest(proxy string, namespace string, version string, isUpgrade bool, registry string, imagePullSecretContent string) (ConnectManifestResponse, error) {
	var result ConnectManifestResponse
	// Create a url object to append parameters to it
	APIURL := prodAddr + "v1beta1/" + c.Resource.Name + ":generateConnectManifest"
	// Create the url parameters
	u, err := url.Parse(APIURL)
	if err != nil {
		return result, fmt.Errorf("Parsing %v url: %w", APIURL, err)
	}
	q := u.Query()
	q.Set("alt", "json")
	q.Set("name", c.Resource.Name)
	if proxy != "" {
		q.Set("connectAgent.proxy", proxy)
	}
	if namespace != "gke-connect" {
		q.Set("connectAgent.namespace", namespace)
	}
	if version != "" {
		q.Set("version", version)
	}
	if isUpgrade {
		q.Set("isUpgrade", "true")
	}
	if registry != "" {
		q.Set("registry", registry)
	}
	if imagePullSecretContent != "" {
		q.Set("imagePullSecretContent", imagePullSecretContent)
	}
	u.RawQuery = q.Encode()
	// Go ahead with the request
	response, err := c.svc.client.Get(u.String())
	if err != nil {
		return result, fmt.Errorf("GET request: %w", err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return result, fmt.Errorf("reading get request body: %w", err)
	}

	statusOK := response.StatusCode >= 200 && response.StatusCode < 300
	if !statusOK {
		return result, fmt.Errorf("Bad status code: %v", string(body))
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, fmt.Errorf("json Un-marshaling body: %w", err)
	}

	return result, nil

}

// ConnectManifestResponse contains the connect agent manifest response
type ConnectManifestResponse struct {
	Manifest []ConnectAgentResource `json:"manifest"`
}

// ConnectAgentResource is part of GenerateConnectManifestResponse
type ConnectAgentResource struct {
	Type     ConnectAgentResourceType `json:"type"`
	Manifest string                   `json:"manifest"`
}

// ConnectAgentResourceType is part of ConnectAgentResource
type ConnectAgentResourceType struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
}
