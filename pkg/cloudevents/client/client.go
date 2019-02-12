package client

import (
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

// Client wraps Builder, and is intended to be configured for a single event
// type and target
type Client struct {
	sender transport.Sender
}

func NewHttpClient(target string) *Client { // , builder Builder
	c := &Client{
		//builder: builder,
		//Target:  target,
	}
	return c
}

func (c *Client) Send(event cloudevents.Event) error {
	//c.sender.Send()
	//
	//context.WithValue()
	//
	//ctx := context.TODO()
	//
	//ctx.
	//
	//
	//req, err := c.builder.Build(c.Target, data, overrides...)
	//if err != nil {
	//	return err
	//}
	//client := &http.Client{}
	//resp, err := client.Do(req)
	//if err != nil {
	//	return err
	//}
	//defer resp.Body.Close()
	//if accepted(resp) {
	//	return nil
	//}
	//return fmt.Errorf("error sending cloudevent: %s", status(resp))
	return nil
}

//
//// accepted is a helper method to understand if the response from the target
//// accepted the CloudEvent.
//func accepted(resp *http.Response) bool {
//	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
//		return true
//	}
//	return false
//}
//
//// status is a helper method to read the response of the target.
//func status(resp *http.Response) string {
//	status := resp.Status
//	body, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		return fmt.Sprintf("Status[%s] error reading response body: %v", status, err)
//	}
//	return fmt.Sprintf("Status[%s] %s", status, body)
//}
