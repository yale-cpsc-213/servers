package questions

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	hwrandom "github.com/yale-cpsc-213/hwutils/random"
)

type serverQuestion func(string, string) (bool, string, error)

func statusText(pass bool) string {
	if pass {
		return "✅ PASS"
	}
	return "❌ FAIL"
}

// TestAll ...
func TestAll(rawURL string, showOutput bool) error {
	doLog := func(args ...interface{}) {
		if showOutput {
			fmt.Println(args...)
		}
	}
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return err
	}

	questions := []serverQuestion{
		indexIsUp,
		protected,
		stringUpperCase,
	}
	for _, question := range questions {
		passed, questionText, err := question(parsedURL.Scheme, parsedURL.Host)
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

func readResponseBody(response *http.Response) (string, error) {
	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	bodyString := string(bodyBytes)
	return bodyString, err
}

func testBodyEquals(response *http.Response, err error, questionText string, expectedBody string) (bool, string, error) {
	if err != nil {
		log.Println("error!")
		return false, questionText, err
	}
	dump, err2 := readResponseBody(response)
	if err2 != nil {
		log.Println("error!")
		return false, questionText, err
	}
	body := strings.Trim(string(dump), " ")
	log.Println("Body =", body)
	if body == expectedBody {
		return true, questionText, nil
	}
	return false, questionText, nil
}

func getAndCheckStatus(scheme string, host string, urlPath string, query url.Values, questionText string, expectedStatus int) (bool, string, error) {
	parsedURL := url.URL{
		Scheme:   scheme,
		Host:     host,
		Path:     urlPath,
		RawQuery: query.Encode(),
	}
	netClient := newClient()
	response, err := netClient.Get(parsedURL.String())
	log.Println(parsedURL.String())
	return testStatusEquals(response, err, questionText, expectedStatus)
}

func getAndCheckBody(scheme string, host string, urlPath string, query url.Values, questionText string, expectedBody string) (bool, string, error) {
	parsedURL := url.URL{
		Scheme:   scheme,
		Host:     host,
		Path:     urlPath,
		RawQuery: query.Encode(),
	}
	netClient := newClient()
	response, err := netClient.Get(parsedURL.String())
	return testBodyEquals(response, err, questionText, expectedBody)
}

func indexIsUp(scheme string, baseURL string) (bool, string, error) {
	return getAndCheckStatus(
		scheme,
		baseURL,
		"/",
		url.Values{},
		"Your website is up (requesting /)",
		http.StatusOK,
	)
}

func protected(scheme string, baseURL string) (bool, string, error) {
	return getAndCheckStatus(
		scheme,
		baseURL,
		"/protected",
		url.Values{},
		"Some parts are protected",
		http.StatusUnauthorized,
	)
}

func stringUpperCase(scheme string, baseURL string) (bool, string, error) {
	questionText := "Strings API converts to uppercase"
	query := url.Values{}
	randomString := hwrandom.LowerString(50)
	query.Set("value", randomString)
	return getAndCheckBody(
		scheme,
		baseURL,
		"/strings/upper",
		query,
		questionText,
		strings.ToUpper(randomString),
	)
}
