package main

import (
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	iplocationApiUrl = "https://api.iplocation.net"
)

var restyClient *resty.Client

func fetchIpLocation(ip string) (IpLocationResponse, error) {
	setupRestyClient()

	ipLoc := &IpLocationResponse{}

	_, err := restyClient.R().
		SetQueryParam("ip", ip).
		SetHeader("Accept", "application/json").
		SetResult(ipLoc).
		Get("/")
	if err != nil {
		return IpLocationResponse{}, err
	}

	return *ipLoc, nil
}

func setupRestyClient() {
	if restyClient == nil {
		restyClient = resty.New()
		restyClient.SetTimeout(time.Second * time.Duration(timeout))
		restyClient.SetBaseURL(iplocationApiUrl)
	}
}

type IpLocationResponse struct {
	IP              string `json:"ip"`
	IPNumber        string `json:"ip_number"`
	IPVersion       int    `json:"ip_version"`
	CountryName     string `json:"country_name"`
	CountryCode2    string `json:"country_code2"`
	Isp             string `json:"isp"`
	ResponseCode    string `json:"response_code"`
	ResponseMessage string `json:"response_message"`
}
