package hub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/MayaraCloud/terraform-provider-anthos/debug"
	"github.com/avast/retry-go"
)

// GetMembership gets details of a hub membership.
// This method also initializes/updates the client component
func (c *Client) GetMembership(membershipID string, checkNotExisting bool) error {
	// Call the gkehub api
	APIURL := prodAddr + "v1/projects/" + c.projectID + "/locations/" + c.location + "/memberships/" + membershipID
	response, err := c.svc.client.Get(APIURL)
	if err != nil {
		return fmt.Errorf("get request: %w", err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("reading get request body: %w", err)
	}

	// If we are checking if the resource does not exist
	// we need a 404 here
	if checkNotExisting && response.StatusCode == 404 {
		return nil
	}

	statusOK := response.StatusCode >= 200 && response.StatusCode < 300
	if !statusOK {
		return fmt.Errorf("Bad %v status code: %v", response.StatusCode, string(body))
	}

	err = json.Unmarshal(body, &c.Resource)
	if err != nil {
		return fmt.Errorf("un-marshaling request body: %w", err)
	}

	if checkNotExisting && response.StatusCode != 404 {
		return fmt.Errorf("The resource already exists in the Hub: %v", string(body))
	}

	return nil
}

// CreateMembership creates a hub membership
// The client object should already contain the
// updated resource component updated in another method
func (c *Client) CreateMembership(membershipID string) error {
	// Validate exclusivity if the cluster has a manifest CRD present
	if c.K8S.CRManifest != "" {
		err := c.ValidateExclusivity(membershipID)
		if err != nil {
			return fmt.Errorf("Validating exclusivity: %w", err)
		}
	}
	// Calling the creation API
	createResponse, err := c.CallCreateMembershipAPI(membershipID)
	if err != nil {
		return fmt.Errorf("Calling CallCreateMembershipAPI: %w", err)
	}

	// Wait until we get an ok from CheckOperation
	retry.Attempts(60)
	err = retry.Do(
		func() error {
			return c.CheckOperation(createResponse["name"].(string))
		})

	if err != nil {
		return fmt.Errorf("Retry checking CreateMembership operation: %w", err)
	}
	return nil
}

// CallCreateMembershipAPI creates a hub membership
// The client object should already contain the
// updated resource component updated in another method
func (c *Client) CallCreateMembershipAPI(membershipID string) (HTTPResult, error) {
	// Create the json POST request body
	var rawBody struct {
		Description string `json:"description"`
		ExternalID  string `json:"externalId"`
	}
	rawBody.Description = c.Resource.Description
	rawBody.ExternalID = c.K8S.UUID

	body, err := json.Marshal(rawBody)
	if err != nil {
		return nil, fmt.Errorf("Marshaling create request body: %w", err)
	}
	// Create a url object to append parameters to it
	APIURL := prodAddr + "v1/projects/" + c.projectID + "/locations/" + c.location + "/memberships"
	u, err := url.Parse(APIURL)
	if err != nil {
		return nil, fmt.Errorf("Parsing %v url: %w", APIURL, err)
	}
	q := u.Query()
	q.Set("alt", "json")
	q.Set("membershipId", membershipID)
	u.RawQuery = q.Encode()
	// Go ahead with the request
	response, err := c.svc.client.Post(u.String(), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create POST request: %w", err)
	}
	defer response.Body.Close()

	return DecodeHTTPResult(response.Body)
}

// HTTPResult is used to store the result of an http request
type HTTPResult map[string]interface{}

// DecodeHTTPResult decodes an http response body
func DecodeHTTPResult(httpBody io.ReadCloser) (HTTPResult, error) {
	var h HTTPResult
	err := json.NewDecoder(httpBody).Decode(&h)
	if err != nil {
		return nil, fmt.Errorf("Decoding http body response: %w", err)
	}
	return h, nil
}

// CheckOperation checks a hub operation status and returns true if the operation is done
func (c *Client) CheckOperation(operationName string) error {
	// Create a url object to append parameters to it
	APIURL := prodAddr + "v1/" + operationName
	// Create the url parameters
	u, err := url.Parse(APIURL)
	if err != nil {
		return retry.Unrecoverable(fmt.Errorf("Parsing %v url: %w", APIURL, err))
	}
	q := u.Query()
	q.Set("alt", "json")
	u.RawQuery = q.Encode()
	// Go ahead with the request
	response, err := c.svc.client.Get(u.String())
	if err != nil {
		return retry.Unrecoverable(fmt.Errorf("GET request: %w", err))
	}
	defer response.Body.Close()

	statusOK := response.StatusCode >= 200 && response.StatusCode < 300
	if !statusOK {
		return retry.Unrecoverable(fmt.Errorf("Bad status code: %v", response.StatusCode))
	}

	result, err := DecodeHTTPResult(response.Body)
	if err != nil {
		return retry.Unrecoverable(fmt.Errorf("Calling DecodeHTTPResult: %w", err))
	}

	if result["done"] == true {
		return nil
	}

	return fmt.Errorf("Failed to check operation: %v", result)
}

// ValidateExclusivity checks the cluster exclusivity against the API
func (c *Client) ValidateExclusivity(membershipID string) error {
	// Call the gkehub api
	APIURL := prodAddr + "v1beta1/projects/" + c.projectID + "/locations/" + c.location + "/memberships:validateExclusivity"
	// Create the url parameters
	u, err := url.Parse(APIURL)
	if err != nil {
		return fmt.Errorf("Parsing %v url: %w", APIURL, err)
	}
	q := u.Query()
	q.Set("crManifest", c.K8S.CRManifest)
	q.Set("intendedMembership", membershipID)
	q.Set("alt", "json")
	u.RawQuery = q.Encode()
	// Go ahead with the request
	response, err := c.svc.client.Get(u.String())
	if err != nil {
		return fmt.Errorf("get request: %w", err)
	}
	defer response.Body.Close()

	statusOK := response.StatusCode >= 200 && response.StatusCode < 300
	if !statusOK {
		return fmt.Errorf("Bad status code: %v", response.Body)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("reading get request body: %w", err)
	}
	var result GRCPResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return fmt.Errorf("json Un-marshaling body: %w", err)
	}

	// 0 == OK in gRCP codes, see below.
	if result.Status.Code != 0 {
		return fmt.Errorf("%v", result.Status.Message)
	}

	return nil
}

// GRCPResponse follows the https://cloud.google.com/apis/design/errors
// Code must be one of the following
// https://github.com/googleapis/googleapis/blob/master/google/rpc/code.proto
type GRCPResponse struct {
	Status GRCPResponseStatus `json:"status"`
}

// GRCPResponseStatus is the inner GRCPResponse struct
type GRCPResponseStatus struct {
	// Code contains the validation result. As such,
	// * OK means that exclusivity may be obtained if the manifest produced by
	// GenerateExclusivityManifest can successfully be applied.
	// * ALREADY_EXISTS means that the Membership CRD is already owned by another
	// Hub. See status.message for more information when this occurs
	Code    int32                  `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details"`
}

// GenerateExclusivity checks the cluster exclusivity against the API
func (c *Client) GenerateExclusivity(membershipID string) error {
	// Call the gkehub api
	APIURL := prodAddr + "v1beta1/projects/" + c.projectID + "/locations/" + c.location + "/memberships/" + membershipID + ":generateExclusivityManifest"

	// Create the url parameters
	u, err := url.Parse(APIURL)
	if err != nil {
		return fmt.Errorf("Parsing %v url: %w", APIURL, err)
	}
	q := u.Query()
	q.Set("name", c.Resource.Name)
	q.Set("crManifest", c.K8S.CRManifest)
	q.Set("crdManifest", c.K8S.CRDManifest)
	q.Set("alt", "json")
	u.RawQuery = q.Encode()
	// Go ahead with the request
	response, err := c.svc.client.Get(u.String())
	if err != nil {
		return fmt.Errorf("get request: %w", err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("reading get request body: %w", err)
	}

	debug.GoLog("GenerateExclusivity: manifest response: " + string(body))

	statusOK := response.StatusCode >= 200 && response.StatusCode < 300
	if !statusOK {
		return fmt.Errorf("Bad status code: %v", string(body))
	}

	type manifestResponse struct {
		CRDManifest string `json:"crdManifest"`
		CRManifest  string `json:"crManifest"`
	}
	var result manifestResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return fmt.Errorf("json Un-marshaling body: %w", err)
	}

	// Populate the client with the manifest and CRD from the gkehub API
	c.K8S.CRDManifest = result.CRDManifest
	c.K8S.CRManifest = result.CRManifest
	if result.CRDManifest == "" && result.CRManifest == "" {
		debug.GoLog("GenerateExclusivity: the client received empty strings")
	}

	return nil
}

// DeleteMembership deletes a hub membership
// The client object should already contain the
// updated resource component updated in another method
func (c *Client) DeleteMembership() error {
	// Delete a url object to append parameters to it
	APIURL := prodAddr + "v1/" + c.Resource.Name

	u, err := url.Parse(APIURL)
	if err != nil {
		return fmt.Errorf("Parsing %v url: %w", APIURL, err)
	}
	q := u.Query()
	q.Set("alt", "json")
	u.RawQuery = q.Encode()
	// Go ahead with the request
	req, err := http.NewRequest("DELETE", u.String(), nil)
	if err != nil {
		return fmt.Errorf("Creating Delete request: %w", err)
	}
	response, err := c.svc.client.Do(req)
	if err != nil {
		return fmt.Errorf("Sending DELETE request: %w", err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("reading get request body: %w", err)
	}

	statusOK := response.StatusCode >= 200 && response.StatusCode < 300
	if !statusOK {
		return fmt.Errorf("Bad %v status code: %v", response.StatusCode, string(body))
	}

	return nil
}
