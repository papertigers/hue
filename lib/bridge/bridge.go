package bridge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/papertigers/hue/lib/config"
)

const (
	// Hue docs say to use "IpBridge" over "hue-bridgeid"
	_SSDPIdentifier    = "IpBridge"
	_DefaultBufferSize = 256
	_DefaultTimeout    = 30 * time.Second
	_DefaultNumBridges = 8
)

var _SSDPData = []string{
	"M-SEARCH * HTTP/1.1",
	"HOST:239.255.255.250:1900",
	"MAN:\"ssdp:discover\"",
	"ST:ssdp:all",
	"MX:1",
}

// Discover Hue bridges via SSDP.
// Returns a map of IP.String() to empty struct.
func Discover() ([]string, error) {
	bridgeSet := make([]string, 0, _DefaultNumBridges)

	rAddr, err := net.ResolveUDPAddr("udp4", "239.255.255.250:1900")
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp4", nil, rAddr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	lAddr, err := net.ListenMulticastUDP("udp4", nil, rAddr)
	if err != nil {
		return nil, err
	}
	defer lAddr.Close()

	// Write discovery packet to network
	if _, err = conn.Write([]byte(strings.Join(_SSDPData, "\r\n"))); err != nil {
		return nil, err
	}

	// Read responses back for short time period
	lAddr.SetReadDeadline(time.Now().Add(_DefaultTimeout))
	var buf bytes.Buffer
	buf.Grow(_DefaultBufferSize)

	for {
		buf.Reset()
		n, addr, err := lAddr.ReadFromUDP(buf.Bytes())
		if err != nil {
			switch osErr := err.(*net.OpError); {
			case osErr.Timeout():
				// Timeout
				return bridgeSet, nil
			case osErr.Temporary():
				// Transient condition
				return nil, err
			default:
				return bridgeSet, err
			}
		}
		if bytes.Contains(buf.Bytes()[:n], []byte(_SSDPIdentifier)) {
			bridgeSet = append(bridgeSet, addr.IP.String())
		}
	}
}

// Bridge represents a Hue Bridge
type Bridge struct {
	IP string
}

func (b *Bridge) CreateUser() (*config.CreateUserResult, error) {
	payload := &config.CreateUser{
		DeviceType: "gohue#papertigers",
	}
	return b.CreateUserWithName(payload)
}

func (b *Bridge) CreateUserWithName(payload *config.CreateUser) (*config.CreateUserResult, error) {
	client := &http.Client{
		Timeout: time.Second * 5,
	}

	method := http.MethodPost
	path := fmt.Sprintf("http://%s/api", b.IP)

	var bodyReq io.ReadSeeker
	marshaled, err := json.MarshalIndent(payload, "", "    ")
	if err != nil {
		return nil, err
	}
	bodyReq = bytes.NewReader(marshaled)

	if err != nil {
		return nil, errwrap.Wrapf("Error creating POST request: {{err}}", err)
	}

	req, err := http.NewRequest(method, path, bodyReq)
	if err != nil {
		return nil, errwrap.Wrapf("Error constructing HTTP request: {{err}}", err)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, errwrap.Wrapf("Error executing HTTP request: {{err}}", err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errwrap.Wrapf("Error reading HTTP body: {{err}}", err)
	}

	var results []config.CreateUserResult
	err = json.Unmarshal(body, &results)
	if err != nil {
		return nil, errwrap.Wrapf("Error unmarshaling CreateUserResult: {{err}}", err)
	}

	return &results[0], nil
}
