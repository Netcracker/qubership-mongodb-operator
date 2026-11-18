// Copyright 2024-2025 NetCracker Technology Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type HttpRequest interface {
	SendRequest(url, method string, bodyStruct interface{}) (*http.Response, error)
	ReadResponseBody(response *http.Response) ([]byte, error)
}

type HttpRequestImpl struct {
	client HttpClient
}

func NewHttpRequest() HttpRequest {
	return &HttpRequestImpl{client: &http.Client{}}
}

func (r *HttpRequestImpl) SendRequest(url, method string, bodyStruct interface{}) (*http.Response, error) {
	var req *http.Request
	var err error

	if method == http.MethodPost {
		codedBody, errm := json.Marshal(bodyStruct)
		if errm != nil {
			return nil, errm
		}
		req, err = http.NewRequest(method, url, bytes.NewReader(codedBody))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		return nil, err
	}

	return r.client.Do(req)
}

func (d *HttpRequestImpl) ReadResponseBody(response *http.Response) ([]byte, error) {
	if response.StatusCode >= 200 && response.StatusCode <= 299 {
		defer response.Body.Close()
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		return body, nil
	} else {
		return nil, errors.New(fmt.Sprintf("Response code is %d", response.StatusCode))
	}
}

type OperatorClient interface {
	GetKeyFile() (string, error)
	GetStatus() (string, error)
}

type OperatorClientImpl struct {
	host   string
	client HttpRequest
}

func NewOperatorClinet(host string) OperatorClient {
	return &OperatorClientImpl{client: NewHttpRequest(), host: host}
}

func (o *OperatorClientImpl) GetKeyFile() (string, error) {
	resp, err := o.client.SendRequest(fmt.Sprintf("%s/%s", o.host, KeyFileURI), http.MethodGet, nil)
	if err != nil {
		return "", err
	}
	result, err := o.client.ReadResponseBody(resp)
	if err != nil {
		return "", err
	}
	var key string
	err = json.Unmarshal(result, &key)
	return key, nil
}

func (o *OperatorClientImpl) GetStatus() (string, error) {
	resp, err := o.client.SendRequest(fmt.Sprintf("%s/%s", o.host, HealthURI), http.MethodGet, nil)
	if err != nil {
		return "", nil
	}

	respBody, err := o.client.ReadResponseBody(resp)
	if err != nil {
		return "", err
	}

	var result map[string]string

	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return "", nil
	}

	return result["status"], nil
}
