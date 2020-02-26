// just-install - The simple package installer for Windows
// Copyright (C) 2020 just-install authors.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, version 3 of the License.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package fetch

import (
	"net"
	"net/http"
	"time"
)

// ConnectionPhaseTimeout is the timeout used as upper bound for various phases of the connection to
// a remote host (i.e. establishing a TCP connection, TLS handshake, etc). This is kept as short as
// possible to immediately catch transient network errors.
const ConnectionPhaseTimeout = 10 * time.Second

// RequestTimeout is the timeout used as the upper bound for an entire HTTP request, including the
// time needed to download the requested file.
const RequestTimeout = 30 * time.Minute

// Transport is an HTTP transport optimized to perform a sigle request to a single host, with short
// timeouts for various connection phases.
var Transport = &http.Transport{
	DialContext: (&net.Dialer{
		DualStack: true,
		KeepAlive: 0,
		Timeout:   ConnectionPhaseTimeout,
	}).DialContext,
	DisableKeepAlives:     true,
	ExpectContinueTimeout: ConnectionPhaseTimeout,
	IdleConnTimeout:       ConnectionPhaseTimeout,
	MaxConnsPerHost:       1,
	Proxy:                 http.ProxyFromEnvironment,
	ResponseHeaderTimeout: ConnectionPhaseTimeout,
	TLSHandshakeTimeout:   ConnectionPhaseTimeout,
}

// NewClient creates a new HTTP client with a default request timeout (see also `RequestTimeout`)
// that uses our `Transport`. Unlike Go stdlib's HTTP client, ours is to be closed and discarded
// after one request.
func NewClient() *http.Client {
	return &http.Client{
		Timeout:   RequestTimeout,
		Transport: Transport,
	}
}
