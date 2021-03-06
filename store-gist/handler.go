package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// Handle a serverless request
func Handle(payload []byte) string {

	if os.Getenv("Http_Method") != "POST" {
		fmt.Fprintf(os.Stderr, "You must post a body to this function.")
		os.Exit(1)
	}

	url := "https://api.github.com/gists"

	var jsonStr = []byte(`{
                "description": "` + fmt.Sprintf("A gist capturing %d bytes", len(payload)) + `",
                "public": true,
                "files": {
                        "post-body.txt": {
                            "content": "` + string(payload) + `"
                        }
                    }
                }`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", "token "+readSecret()) // The token
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	if resp.Status == "201 Created" {
		res, getErr := http.Get(resp.Header.Get("Location"))
		if getErr != nil {
			fmt.Fprintf(os.Stderr, getErr.Error())
			os.Exit(1)
		}

		bytesOut, _ := ioutil.ReadAll(res.Body)
		gistResult := GistResult{}
		json.Unmarshal(bytesOut, &gistResult)
		return gistResult.HtmlURL
	}

	resBody, _ := ioutil.ReadAll(resp.Body)

	fmt.Fprintf(os.Stderr, fmt.Sprintf("Couldn't create file %d %s\n", resp.StatusCode, string(resBody)))
	os.Exit(1)

	return ""
}

type GistResult struct {
	HtmlURL string `json:"html_url"`
}

func readSecret() string {
	val, err := ioutil.ReadFile("/var/openfaas/secrets/github-token")
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	return strings.TrimSpace(string(val))
}
