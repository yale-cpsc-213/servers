package questions

import (
	"fmt"
	"net/http"
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

	questions := []serverQuestion{indexIsUp}
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

func indexIsUp(url string) (bool, string, error) {
	testDesc := fmt.Sprintf("Your website is up (requesting %s)", url)
	netClient := newClient()
	response, err := netClient.Get(url)
	if err != nil {
		return false, testDesc, err
	}
	if response.StatusCode == http.StatusOK {
		return true, testDesc, nil
	}
	return false, testDesc, nil
}
