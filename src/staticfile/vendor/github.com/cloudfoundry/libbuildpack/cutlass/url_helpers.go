package cutlass

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/cloudfoundry/libbuildpack/cutlass/models"
)

func Get(app models.CfApp, path string, headers map[string]string) (string, map[string][]string, error) {
	url, err := app.GetUrl(path)
	if err != nil {
		return "", map[string][]string{}, err
	}
	client := &http.Client{}
	if headers["NoFollow"] == "true" {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
		delete(headers, "NoFollow")
	}
	req, _ := http.NewRequest("GET", url, nil)
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	if headers["user"] != "" && headers["password"] != "" {
		req.SetBasicAuth(headers["user"], headers["password"])
		delete(headers, "user")
		delete(headers, "password")
	}
	fmt.Println("Get 0:", req.URL)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Get 1:", err)
		return "", map[string][]string{}, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Get 2:", err)
		return "", map[string][]string{}, err
	}
	resp.Header["StatusCode"] = []string{strconv.Itoa(resp.StatusCode)}
	fmt.Println("Get 3:", string(data), resp.Header, err)
	return string(data), resp.Header, err
}
func GetBody(app models.CfApp, path string) (string, error) {
	body, _, err := Get(app, path, map[string]string{})
	// TODO: Non 200 ??
	// if !(len(headers["StatusCode"]) == 1 && headers["StatusCode"][0] == "200") {
	// 	return "", fmt.Errorf("non 200 status: %v", headers)
	// }
	return body, err
}
