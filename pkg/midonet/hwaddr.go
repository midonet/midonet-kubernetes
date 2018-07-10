// Copyright 2015 flannel authors
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

package midonet

import (
	"fmt"
	"net"
)

// delivered from flannel/backend/vxlan/vxlan.go

// So we can make it JSON (un)marshalable
type HardwareAddr net.HardwareAddr

func (hw HardwareAddr) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", net.HardwareAddr(hw))), nil
}

func (hw *HardwareAddr) UnmarshalJSON(bytes []byte) error {
	if len(bytes) < 2 || bytes[0] != '"' || bytes[len(bytes)-1] != '"' {
		return fmt.Errorf("error parsing hardware addr")
	}

	bytes = bytes[1 : len(bytes)-1]

	mac, err := net.ParseMAC(string(bytes))
	if err != nil {
		return err
	}

	*hw = HardwareAddr(mac)
	return nil
}

func (hw HardwareAddr) String() string {
	return net.HardwareAddr(hw).String()
}
