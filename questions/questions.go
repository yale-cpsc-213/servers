package questions

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"golang.org/x/net/html"

	hwrandom "github.com/yale-cpsc-213/hwutils/random"
	hwstrings "github.com/yale-cpsc-213/hwutils/strings"
)

type serverQuestion func(string, string) (bool, string, error)
type responseTester func(*http.Response) (bool, error)

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
		stringReverse,
		stringConcatenate,
		// movieIndex,
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
		return false, questionText, err
	}
	dump, err2 := readResponseBody(response)
	if err2 != nil {
		return false, questionText, err
	}
	body := strings.Trim(string(dump), " ")
	if body == expectedBody {
		return true, questionText, nil
	}
	return false, questionText, nil
}

func testResponse(response *http.Response, err error, questionText string, testFunc responseTester) (bool, string, error) {

	result, err := testFunc(response)
	if result && err == nil {
		return true, questionText, nil
	}
	return false, questionText, nil
}

func getAndCheckFunction(scheme string, host string, urlPath string, query url.Values, questionText string, testFunc responseTester) (bool, string, error) {
	parsedURL := url.URL{
		Scheme:   scheme,
		Host:     host,
		Path:     urlPath,
		RawQuery: query.Encode(),
	}
	netClient := newClient()
	response, err := netClient.Get(parsedURL.String())
	return testResponse(response, err, questionText, testFunc)
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
	return testStatusEquals(response, err, questionText, expectedStatus)
}

func getAndCheckBody(scheme string, host string, urlPath string, query url.Values, questionText string, expectedBody string) (bool, string, error) {
	testFunc := func(response *http.Response) (bool, error) {
		body, err := readResponseBody(response)
		if err != nil {
			return false, err
		}
		if body == expectedBody {
			return true, nil
		}
		return false, nil
	}
	return getAndCheckFunction(
		scheme,
		host,
		urlPath,
		query,
		questionText,
		testFunc,
	)
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
	questionText := "Strings API can convert to uppercase"
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

func stringReverse(scheme string, baseURL string) (bool, string, error) {
	questionText := "Strings API can reverse strings"
	query := url.Values{}
	randomString := hwrandom.LowerString(50)
	query.Set("value", randomString)
	return getAndCheckBody(
		scheme,
		baseURL,
		"/strings/reverse",
		query,
		questionText,
		hwstrings.Reverse(randomString),
	)
}
func stringConcatenate(scheme string, baseURL string) (bool, string, error) {
	questionText := "Strings API can concatenate"
	query := url.Values{}
	randomString := hwrandom.LowerString(50)
	query.Set("value", randomString)
	times := rand.Intn(5) + 1
	query.Set("times", strconv.Itoa(times))
	expectedBody := ""
	for i := 0; i < times; i++ {
		expectedBody += randomString
	}
	return getAndCheckBody(
		scheme,
		baseURL,
		"/strings/concatenate",
		query,
		questionText,
		expectedBody,
	)
}

func debugHTML(n *html.Node) {
	var buf bytes.Buffer
	if err := html.Render(&buf, n); err != nil {
		log.Fatalf("Render error: %s", err)
	}
	fmt.Println(buf.String())
}

func movieIndex(scheme string, baseURL string) (bool, string, error) {
	hasMovies := func(response *http.Response) (bool, error) {
		doc, err := goquery.NewDocumentFromReader(response.Body)
		if err != nil {
			return false, err
		}
		numLi := doc.Find("li.movie").Length()
		numberOfMoviesExpected := 50
		if numLi == numberOfMoviesExpected {
			return true, nil
		}
		return false, nil
	}
	return getAndCheckFunction(
		scheme,
		baseURL,
		"/movies",
		url.Values{},
		"There is a movie index page",
		hasMovies,
	)
}
