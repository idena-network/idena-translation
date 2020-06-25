package node

import (
	"encoding/json"
	"fmt"
	"github.com/idena-network/idena-translation/types"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type Client interface {
	GetSignatureAddress(value, signature string) (string, error)
	IsIdentity(address string) (bool, error)
}

type Response struct {
	Result interface{} `json:"result,omitempty"`
	Error  *RespError  `json:"error,omitempty"`
}

type RespError struct {
	Message string `json:"message"`
}

type Identity struct {
	State string `json:"state"`
}

func NewClient(apiUrl string) Client {
	return &clientImpl{
		apiUrl: apiUrl,
	}
}

type clientImpl struct {
	apiUrl string
}

func (c *clientImpl) GetSignatureAddress(value, signature string) (string, error) {
	urlValues := url.Values{}
	urlValues.Add("value", value)
	urlValues.Add("signature", signature)
	responseBytes, err := sendRequest(fmt.Sprintf("%v/api/SignatureAddress?%v", c.apiUrl, urlValues.Encode()))
	if err != nil {
		return "", err
	}
	var address string
	var response = Response{
		Result: &address,
	}
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return "", err
	}
	if response.Error != nil {
		return "", &types.BadRequestError{
			Message: response.Error.Message,
		}
	}
	return address, nil
}

func (c *clientImpl) IsIdentity(address string) (bool, error) {
	responseBytes, err := sendRequest(fmt.Sprintf("%v/api/identity/%v", c.apiUrl, address))
	if err != nil {
		return false, err
	}
	var identity Identity
	var response = Response{
		Result: &identity,
	}
	if err := json.Unmarshal(responseBytes, &response); err != nil {
		return false, err
	}
	if response.Error != nil {
		if response.Error.Message == "no data found" {
			return false, nil
		}
		return false, errors.New(response.Error.Message)
	}
	return isIdentity(identity.State), nil
}

func sendRequest(req string) ([]byte, error) {
	httpReq, err := http.NewRequest("GET", req, nil)
	if err != nil {
		return nil, err
	}
	var resp *http.Response
	defer func() {
		if resp == nil || resp.Body == nil {
			return
		}
		resp.Body.Close()
	}()
	httpClient := &http.Client{
		Timeout: time.Second * 5,
	}
	resp, err = httpClient.Do(httpReq)
	if err == nil && resp.StatusCode != http.StatusOK {
		err = errors.New(fmt.Sprintf("resp code %v", resp.StatusCode))
	}
	if err != nil {
		return nil, err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read resp")
	}
	return respBody, nil
}

func isIdentity(state string) bool {
	return state == "Newbie" || state == "Verified" || state == "Human"
}
