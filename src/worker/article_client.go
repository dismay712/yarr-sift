package worker

import (
	"errors"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"syscall"
	"time"
)

func validateArticleURL(u *url.URL) error {
	if (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil {
		return errors.New("article URL must be HTTP(S) with a host and no credentials")
	}
	return nil
}

func newArticleClient(timeout time.Duration) *http.Client {
	if timeout <= 0 || timeout > articleTimeout {
		timeout = articleTimeout
	}
	dialer := &net.Dialer{Timeout: 10 * time.Second, Control: checkArticleAddress}
	return &http.Client{
		Timeout: timeout,
		// Direct connections let the dialer check the actual resolved destination,
		// including every redirected connection. A proxy would resolve it elsewhere.
		Transport: &http.Transport{
			DialContext:            dialer.DialContext,
			ForceAttemptHTTP2:      true,
			MaxIdleConns:           32,
			IdleConnTimeout:        time.Minute,
			TLSHandshakeTimeout:    10 * time.Second,
			MaxResponseHeaderBytes: 1 << 20,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return errors.New("too many article redirects")
			}
			return validateArticleURL(req.URL)
		},
	}
}

// These special-use ranges pass netip's global-unicast and private-address tests.
var nonPublicRanges = [...]netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("192.88.99.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("2001::/23"),
	netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("2002::/16"),
	netip.MustParsePrefix("3fff::/20"),
}

var globalIPv6 = netip.MustParsePrefix("2000::/3")

// net.Dialer invokes Control with a numeric destination after DNS resolution
// and before connecting, avoiding a second DNS lookup after validation.
func checkArticleAddress(_, address string, _ syscall.RawConn) error {
	dest, err := netip.ParseAddrPort(address)
	if err != nil {
		return err
	}
	ip := dest.Addr().Unmap()
	if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.Zone() != "" || (ip.Is6() && !globalIPv6.Contains(ip)) {
		return errors.New("non-public article address")
	}
	for _, prefix := range nonPublicRanges {
		if prefix.Contains(ip) {
			return errors.New("non-public article address")
		}
	}
	return nil
}
