package questions

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"
)

type serverQuestion func(string) (bool, string, error)

func statusText(pass bool) string {
	if pass {
		return "✅ PASS"
	}
	return "❌ FAIL"
}

// TestAll ...
func TestAll(url string, showOutput bool) error {
	doLog := func(args ...interface{}) {
		if showOutput {
			fmt.Println(args...)
		}
	}

	questions := []serverQuestion{indexIsUp, protected}
	for _, question := range questions {
		passed, questionText, err := question(url)
		doLog(statusText(passed && (err == nil)), "-", questionText)
	}
	return nil
}

func newClient() *http.Client {
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}
	return netClient
}

func testStatusEquals(response *http.Response, err error, questionText string, expectedStatus int) (bool, string, error) {
	if err != nil {
		return false, questionText, err
	}
	if response.StatusCode == expectedStatus {
		return true, questionText, nil
	}
	return false, questionText, nil
}

func getAndCheckStatus(baseURL string, urlPath string, questionText string, expectedStatus int) (bool, string, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return false, questionText, err
	}
	parsedURL.Path = path.Join(parsedURL.Path, urlPath)
	netClient := newClient()
	response, err := netClient.Get(parsedURL.String())
	return testStatusEquals(response, err, questionText, expectedStatus)

}

func indexIsUp(baseURL string) (bool, string, error) {
	return getAndCheckStatus(
		baseURL,
		"/",
		"Your website is up (requesting /)",
		http.StatusOK,
	)
}
func protected(baseURL string) (bool, string, error) {
	return getAndCheckStatus(
		baseURL,
		"/protected",
		"Some parts are protected",
		http.StatusUnauthorized,
	)
}
