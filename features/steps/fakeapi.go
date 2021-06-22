package steps

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/maxcnunes/httpfake"
)

// FakeAPI contains all the information for a fake component API
type FakeAPI struct {
	fakeHTTP                     *httpfake.HTTPFake
	outboundRequests             []string
	collectOutboundRequestBodies httpfake.CustomAssertor
}

// NewFakeAPI creates a new fake component API
func NewFakeAPI(t testing.TB) *FakeAPI {
	fa := &FakeAPI{
		fakeHTTP: httpfake.New(httpfake.WithTesting(t)),
	}

	fa.collectOutboundRequestBodies = func(r *http.Request) error {
		// inspect request
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return fmt.Errorf("error reading the outbound request body: %s", err.Error())
		}
		fa.outboundRequests = append(fa.outboundRequests, string(body))
		return nil
	}

	return fa
}

func (f *FakeAPI) setJSONResponseForGet(url string, responseBody string) {
	f.fakeHTTP.NewHandler().Get(url).AssertHeaders("Content-Type").Reply(200).SetHeader("Content-Type", "application/json").Body([]byte(responseBody))
}

func (f *FakeAPI) setJSONResponseForPost(url string, responseBody string) *httpfake.Request {
	request := f.fakeHTTP.NewHandler().Post(url).AssertHeaders("Content-Type")

	request.Reply(200).SetHeader("Content-Type", "application/json").Body([]byte(responseBody))

	return request
}

// Close closes the fake API
func (f *FakeAPI) Close() {
	f.fakeHTTP.Close()
}

// Reset resets the fake API
func (f *FakeAPI) Reset() {
	f.fakeHTTP.Reset()
}
