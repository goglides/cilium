// Copyright 2018 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build privileged_tests

package ipsec

import (
	"bytes"
	"net"
	"os"
	"testing"

	. "gopkg.in/check.v1"

	"github.com/vishvananda/netlink"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type IPSecSuitePrivileged struct{}

var _ = Suite(&IPSecSuitePrivileged{})

var (
	path           = "ipsec_keys_test"
	keysDat        = []byte("hmac(sha256) 0123456789abcdef0123456789abcdef cbc(aes) 0123456789abcdef0123456789abcdef\nhmac(sha256) 0123456789abcdef0123456789abcdef cbc(aes) 0123456789abcdef0123456789abcdef foobar\n")
	invalidKeysDat = []byte("test abcdefghijklmnopqrstuvwzyzABCDEF test abcdefghijklmnopqrstuvwzyzABCDEF\n")
)

func (p *IPSecSuitePrivileged) TestLoadKeysNoFile(c *C) {
	err := LoadIPSecKeysFile(path)
	c.Assert(os.IsNotExist(err), Equals, true)
}

func (p *IPSecSuitePrivileged) TestInvalidLoadKeys(c *C) {
	keys := bytes.NewReader(invalidKeysDat)
	err := loadIPSecKeys(keys)
	spi := 1
	c.Assert(err, NotNil)

	_, local, err := net.ParseCIDR("1.1.3.4/16")
	c.Assert(err, IsNil)
	_, remote, err := net.ParseCIDR("1.2.3.4/16")
	c.Assert(err, IsNil)

	err = UpsertIPSecEndpoint(local, remote, spi, IPSecDirBoth)
	c.Assert(err, NotNil)
}

func (p *IPSecSuitePrivileged) TestLoadKeys(c *C) {
	keys := bytes.NewReader(keysDat)
	err := loadIPSecKeys(keys)
	c.Assert(err, IsNil)
}

func (p *IPSecSuitePrivileged) TestUpsertIPSecEquals(c *C) {
	spi := 1

	_, local, err := net.ParseCIDR("1.2.3.4/16")
	c.Assert(err, IsNil)
	_, remote, err := net.ParseCIDR("1.2.3.4/16")
	c.Assert(err, IsNil)

	key := &ipSecKey{
		Spi:   1,
		ReqID: 1,
		Auth:  &netlink.XfrmStateAlgo{Name: "hmac(sha256)", Key: []byte("0123456789abcdef0123456789abcdef")},
		Crypt: &netlink.XfrmStateAlgo{Name: "cbc(aes)", Key: []byte("0123456789abcdef0123456789abcdef")},
	}

	ipSecKeysGlobal["1.2.3.4"] = key
	ipSecKeysGlobal[""] = key

	err = UpsertIPSecEndpoint(local, remote, spi, IPSecDirBoth)
	c.Assert(err, IsNil)

	err = DeleteIPSecEndpoint(remote.IP, local.IP)
	c.Assert(err, IsNil)

	ipSecKeysGlobal["1.2.3.4"] = nil
	ipSecKeysGlobal[""] = nil

}

func (p *IPSecSuitePrivileged) TestUpsertIPSecEndpoint(c *C) {
	spi := 1

	_, local, err := net.ParseCIDR("1.1.3.4/16")
	c.Assert(err, IsNil)
	_, remote, err := net.ParseCIDR("1.2.3.4/16")
	c.Assert(err, IsNil)

	key := &ipSecKey{
		Spi:   1,
		ReqID: 1,
		Auth:  &netlink.XfrmStateAlgo{Name: "hmac(sha256)", Key: []byte("0123456789abcdef0123456789abcdef")},
		Crypt: &netlink.XfrmStateAlgo{Name: "cbc(aes)", Key: []byte("0123456789abcdef0123456789abcdef")},
	}

	ipSecKeysGlobal["1.1.3.4"] = key
	ipSecKeysGlobal["1.2.3.4"] = key
	ipSecKeysGlobal[""] = key

	err = UpsertIPSecEndpoint(local, remote, spi, IPSecDirBoth)
	c.Assert(err, IsNil)

	err = DeleteIPSecEndpoint(remote.IP, local.IP)
	c.Assert(err, IsNil)

	ipSecKeysGlobal["1.1.3.4"] = nil
	ipSecKeysGlobal["1.2.3.4"] = nil
	ipSecKeysGlobal[""] = nil
}

func (p *IPSecSuitePrivileged) TestUpsertIPSecKeyMissing(c *C) {
	spi := 1

	_, local, err := net.ParseCIDR("1.1.3.4/16")
	c.Assert(err, IsNil)
	_, remote, err := net.ParseCIDR("1.2.3.4/16")
	c.Assert(err, IsNil)

	err = UpsertIPSecEndpoint(local, remote, spi, IPSecDirBoth)
	c.Assert(err, ErrorMatches, "unable to replace local state: IPSec key missing")

	err = DeleteIPSecEndpoint(remote.IP, local.IP)
	c.Assert(err, IsNil)
}
