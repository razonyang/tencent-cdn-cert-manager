package acme

import (
	"fmt"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/providers/dns/alidns"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/providers/dns/tencentcloud"
)

func newProvider(name string) (p challenge.Provider, err error) {
	switch name {
	case "cloudflare":
		p, err = cloudflare.NewDNSProvider()
	case "alidns":
		p, err = alidns.NewDNSProvider()
	case "tencentcloud":
		p, err = tencentcloud.NewDNSProvider()
	default:
		err = fmt.Errorf("unsupported DNS provider %q", name)
	}

	return
}
