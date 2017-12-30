package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"github.com/aprice/observatory"
)

// ErrNotModified is returned when a requested object was not modified.
var ErrNotModified = errors.New("Requested object not modified.")

// ErrAllEndpointsFailed is returned from mutli-endpoint methods when none responded successfully.
var ErrAllEndpointsFailed = errors.New("All endpoints failed. See log for individual errors.")

var userAgent = fmt.Sprintf("Observatory/%s (%s; %s) %s", observatory.Version, runtime.GOOS, runtime.GOARCH, runtime.Version())

var client = new(http.Client)

// SendObject executes a PUT or POST request with the given payload as JSON. It will attempt
// the given endpoints at random until one succeeds or all fail.
func SendObject(method, path string, endpoints []string, payload interface{}) error {
	if len(endpoints) == 0 {
		return errors.New("No endpoints provided.")
	}
	var t1, t2 time.Time
	for _, idx := range rand.Perm(len(endpoints)) {
		url := fmt.Sprintf("http://%s%s", endpoints[idx], path)
		t1 = time.Now()
		err := send(method, url, payload)
		t2 = time.Now()
		if err == nil {
			log.Printf("Saved object to %s in %v", url, t2.Sub(t1))
			return nil
		}
		log.Printf("%s => %s\n", url, err)
	}

	return ErrAllEndpointsFailed
}

func send(method, url string, payload interface{}) error {
	body, err := json.Marshal(&payload)
	if err != nil {
		log.Fatal(err)
	}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if resp != nil {
		defer func() {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}()
	}
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("%s returned %s: %s", url, resp.Status, body)
	}

	return nil
}

// GetObject executes a GET request and returns the payload as JSON. It will attempt
// the given endpoints at random until one succeeds or all fail. It returns the
// expiration time of the payload, if any.
func GetObject(path string, headers map[string]string, endpoints []string, payload interface{}) (expires time.Time, err error) {
	expires = time.Now()
	if len(endpoints) == 0 {
		err = errors.New("No endpoints provided.")
		return
	}

	var t1, t2 time.Time
	for _, idx := range rand.Perm(len(endpoints)) {
		url := fmt.Sprintf("http://%s%s", endpoints[idx], path)
		t1 = time.Now()
		expires, err = get(url, headers, payload)
		t2 = time.Now()
		log.Printf("Received object from %s in %v", url, t2.Sub(t1))
		if err == nil || err == ErrNotModified {
			return expires, nil
		}
		log.Printf("%s => %s\n", url, err)
	}

	err = ErrAllEndpointsFailed
	return
}

func get(url string, headers map[string]string, payload interface{}) (expires time.Time, err error) {
	expires = time.Now()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if resp != nil {
		defer func() {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}()
	}
	if err != nil {
		return
	}
	if resp.StatusCode >= 400 {
		err = fmt.Errorf("%d: %s", resp.StatusCode, resp.Status)
		return
	}

	expireRaw := resp.Header.Get("Expires")
	expires, err = time.Parse(http.TimeFormat, expireRaw)
	if err != nil {
		log.Print(err)
	}

	if resp.StatusCode == 304 {
		err = ErrNotModified
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	log.Printf("[DEBUG] Config: %s\n", string(body))
	err = json.Unmarshal(body, &payload)
	return
}
