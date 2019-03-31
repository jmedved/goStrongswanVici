package goStrongswanVici

import (
	"fmt"
)

type Connection struct {
	ConnConf map[string]IKEConf `json:"connections"`
}

type IKEConf struct {
	LocalAddrs  []string               `json:"local_addrs,omitempty"`
	RemoteAddrs []string               `json:"remote_addrs,omitempty"`
	Proposals   []string               `json:"proposals,omitempty"`
	Version     string                 `json:"version,omitempty"` // 1 for ikev1, 0 for ikev1 & ikev2
	Encap       string                 `json:"encap,omitempty"`   // yes,no
	KeyingTries string                 `json:"keyingtries,omitempty"`
	RekeyTime   string                 `json:"rekey_time,omitempty"`
	ReauthTime  string                 `json:"reauth_time,omitempty"`
	DPDDelay    string                 `json:"dpd_delay,omitempty"`
	Aggressive  string                 `json:"aggressive,omitempty"`
	Vips        string                 `json:"vips,omitempty"`
	Mobike      string                 `json:"mobike,omitempty"`
	SendCertreq string                 `json:"send_certreq,omitempty"`
	LocalAuth   AuthConf               `json:"local,omitempty"`
	RemoteAuth  AuthConf               `json:"remote,omitempty"`
	Pools       []string               `json:"pools,omitempty"`
	Children    map[string]ChildSAConf `json:"children,omitempty"`
}

type AuthConf struct {
	ID         string   `json:"id,omitempty"`
	Round      string   `json:"round,omitempty"`
	AuthMethod string   `json:"auth"` // (psk|pubkey)
	EapId      string   `json:"eap_id,omitempty"`
	Certs      []string `json:"certs,omitempty"`
}

type ChildSAConf struct {
	Local_ts      []string `json:"local_ts,omitempty"`
	Remote_ts     []string `json:"remote_ts,omitempty"`
	ESPProposals  []string `json:"esp_proposals,omitempty"` // aes128-sha1_modp1024
	StartAction   string   `json:"start_action,omitempty"`  // none,trap,start
	CloseAction   string   `json:"close_action,omitempty"`
	ReqID         string   `json:"reqid,omitempty"`
	RekeyTime     string   `json:"rekey_time,omitempty"`
	RekeyBytes    string   `json:"rekey_bytes,omitempty"`
	RekeyPackets  string   `json:"rekey_packets,omitempty"`
	ReplayWindow  string   `json:"replay_window,omitempty"`
	Mode          string   `json:"mode,omitempty"`
	InstallPolicy string   `json:"policies,omitempty"`
	UpDown        string   `json:"updown,omitempty"`
	Priority      string   `json:"priority,omitempty"`
	MarkIn        string   `json:"mark_in,omitempty"`
	MarkOut       string   `json:"mark_out,omitempty"`
	DpdAction     string   `json:"dpd_action,omitempty"`
	LifeTime      string   `json:"life_time,omitempty"`
}

func (c *ClientConn) LoadConn(conn *map[string]*IKEConf) error {
	requestMap := &map[string]interface{}{}

	err := ConvertToGeneral(conn, requestMap)
	if err != nil {
		return fmt.Errorf("error creating request: %#v", err)
	}

	fmt.Printf("\n**** general vici request: %#v\n", *requestMap)

	msg, err := c.Request("load-conn", *requestMap)
	if err != nil {
		return fmt.Errorf("error sending request: %s", err)
	}

	if msg["success"] != "yes" {
		return fmt.Errorf("unsuccessful LoadConn: %s", msg["errmsg"])
	}

	return nil
}
