package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Client interface {
	Query(bucket string) ([]byte, error)
	Del(bucket, key string) error
	Close() error
}

type client struct {
	host   string
	client *http.Client
}

func New(host string) Client {
	return &client{
		host:   host,
		client: &http.Client{},
	}
}

func (c *client) Query(bucket string) ([]byte, error) {
	resp, err := c.client.Get(c.quri(bucket))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (c *client) Del(bucket, key string) error {
	req, err := http.NewRequest("DELETE", c.dUri(bucket, key), nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return err
	}

	return nil
}

func (c *client) Close() error {
	c.client = nil
	return nil
}

func (c *client) quri(bucket string) string {
	return "http://" + c.host + "/buckets/" + url.QueryEscape(bucket) + "/keys?keys=true"
}

func (c *client) dUri(bucket, key string) string {
	return "http://" + c.host + "/buckets/" + url.QueryEscape(bucket) + "/keys/" + url.QueryEscape(key)
}

var (
	buckets = []string{"test", "test2"}
	data    = make(map[string][]string)
)

const host = `127.0.0.1:8098`

func main() {
	type task struct {
		data   []string
		bucket string
	}
	ch := make(chan *task)
	finish := make(chan bool)

	c := New(host)
	defer func() {
		if err := c.Close(); err != nil {
			fmt.Println("Close: ", err.Error())
			return
		}
		finish <- true
	}()

	go func() {
		for _, bucket := range buckets {
			body, err := c.Query(bucket)
			if err != nil {
				fmt.Println("Query: ", err.Error())
				return
			}

			if err = json.Unmarshal(body, &data); err != nil {
				fmt.Println("Unmarshal: ", err.Error())
				return
			}

			if _, ok := data["keys"]; !ok {
				fmt.Println("buckets is not exist")
				return
			}

			if len(data["keys"]) == 0 {
				fmt.Println(bucket, "buckets is nil")
				continue
			}
			ch <- &task{
				data:   data["keys"],
				bucket: bucket,
			}
		}
	}()

	for {
		select {
		case <-finish:
			return
		case ch := <-ch:
			for _, key := range ch.data {
				go func(bucket, key string) {
					if err := c.Del(ch.bucket, key); err != nil {
						panic("Del: " + err.Error())
						return
					}
				}(ch.bucket, key)
			}
		}
	}
}
