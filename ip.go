package kuu

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// IPInfo
type IPInfo struct {
	IP        string `json:"ip"`
	Country   string `json:"country"`
	CountryID string `json:"country_id"`
	Area      string `json:"area"`
	Region    string `json:"region"`
	City      string `json:"city"`
	ISP       string `json:"isp"`
}

// GetIPInfo
func GetIPInfo(ip string) (*IPInfo, error) {
	var (
		err  error
		body struct {
			Code int64
			Data IPInfo
		}
	)
	resp, err := http.Get(fmt.Sprintf("http://ip.taobao.com/service/getIpInfo.php?ip=%s", ip))
	if err == nil && resp != nil {
		var data []byte
		if data, err = ioutil.ReadAll(resp.Body); err == nil {
			err = json.Unmarshal(data, &body)
		}
	}
	return &body.Data, err
}
