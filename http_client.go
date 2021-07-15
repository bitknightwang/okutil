package okutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	JsonContentType = "application/json"
)

func CreateHttpClient() *http.Client {
	// this is overkill
	return &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConns:          10,
			IdleConnTimeout:       30 * time.Second,
			DisableCompression:    true,
		},
		Timeout: 60 * time.Second,
	}
}

func HttpGet(client *http.Client, url string, headers map[string]interface{}) (interface{}, error) {
	if len(url) == 0 {
		return nil, fmt.Errorf("empty url specified")
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if len(headers) > 0 {
		for key, val := range headers {
			req.Header.Set(key, fmt.Sprintf("%s", val))
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get %v got %v", url, resp.Status)
	}

	var result interface{}
	if strings.Contains(resp.Header.Get("content-type"), JsonContentType) && len(bodyBytes) > 0 {
		// Convert response bodyBytes to struct
		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			return nil, err
		}
	} else {
		// as plain string
		result = string(bodyBytes)
	}

	Debugf("response body:\n%v\n", result)

	return result, nil
}

func HttpPostJson(client *http.Client, url string, data interface{}, headers map[string]interface{}) (interface{}, error) {
	if len(url) == 0 {
		return nil, fmt.Errorf("empty url specified")
	}

	var req *http.Request
	var err error
	if data == nil {
		req, err = http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			return nil, err
		}
	} else {
		jsonPayload, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		Debugf("post data:\n" + string(jsonPayload))

		req, err = http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonPayload))
		if err != nil {
			return nil, err
		}
	}


	if len(headers) > 0 {
		for key, val := range headers {
			req.Header.Set(key, fmt.Sprintf("%s", val))
		}
	}
	// force json request
	req.Header.Set("Content-Type", JsonContentType)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to post %v got %v", url, resp.Status)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result interface{}
	if strings.Contains(resp.Header.Get("content-type"), JsonContentType) && len(bodyBytes) > 0 {
		// Convert response bodyBytes to struct
		if err = json.Unmarshal(bodyBytes, &result); err != nil {
			return nil, err
		}
	} else {
		// as plain string
		result = string(bodyBytes)
	}
	Debugf("response body:\n%v\n", result)

	return result, nil
}
