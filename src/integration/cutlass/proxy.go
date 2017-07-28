package cutlass

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
)

func NewProxy() (*httptest.Server, error) {
	addr, err := publicIP()
	if err != nil {
		return nil, err
	}
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := fmt.Sprintf("https://%s%s", r.Host, r.URL)
		resp, err := http.Get(url)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "ERROR", err)
			return
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
		resp.Body.Close()
	}))
	ts.Listener.Close()
	ts.Listener, err = net.Listen("tcp", addr+":0")
	if err != nil {
		return nil, err
	}

	ts.Start()
	return ts, nil
}

func publicIP() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	var addr string
	for _, i := range interfaces {
		if strings.Contains(i.Flags.String(), "up") {
			addrs, err := i.Addrs()
			if err == nil && len(addrs) > 0 {
				addr = addrs[0].String()
			}
		}
	}
	idx := strings.Index(addr, "/")
	if idx > -1 {
		addr = addr[:idx]
	}

	if addr == "" {
		return "", fmt.Errorf("Could not determine IP address")
	}

	return addr, nil
}
