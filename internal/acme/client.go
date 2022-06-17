package acme

import (
	"os"
	"path"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"github.com/razonyang/tencent-cdn-cert-manager/internal/helper"
	"github.com/razonyang/tencent-cdn-cert-manager/internal/letsencrypt"
)

type Client struct {
	*lego.Client
	userRepo UserRepository
	certRepo CertRepository
}

func NewClient(email, dnsProvider string) (c *Client, err error) {
	wd, _ := os.Getwd()
	root := path.Join(wd, "data")
	caDirURL := helper.Getenv("CA_DIR_URL", letsencrypt.CA_DIR_URL_STAGING)
	userRepo := NewFilesystemUserRepository(root, caDirURL)
	user, err := userRepo.Get(email)
	if err != nil {
		return
	}
	config := lego.NewConfig(user)

	config.CADirURL = caDirURL
	config.Certificate.KeyType = certcrypto.RSA2048

	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(config)
	if err != nil {
		return
	}

	provider, err := newProvider(dnsProvider)
	if err != nil {
		return
	}
	err = client.Challenge.SetDNS01Provider(provider)
	if err != nil {
		return
	}

	certRepo := NewFilesystemCertRepository(root, caDirURL)

	// register the new users.
	if user.GetRegistration() == nil {
		var reg *registration.Resource
		reg, err = client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err != nil {
			return
		}
		if u, ok := user.(*User); ok {
			u.Registration = reg
			if err = userRepo.Save(u); err != nil {
				return
			}
		}
	}
	c = &Client{
		Client:   client,
		userRepo: userRepo,
		certRepo: certRepo,
	}
	return
}

func (c *Client) ObtainCertificate(domain string) (cert *certificate.Resource, err error) {
	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}
	cert, err = c.Certificate.Obtain(request)
	if err != nil {
		return
	}
	if err = c.certRepo.Save(cert); err != nil {
		return
	}

	return
}

func (c *Client) RenewCertificate(domain string) (cert *certificate.Resource, err error) {
	cert, err = c.certRepo.Get(domain)
	if err != nil {
		return
	}
	cert, err = c.Client.Certificate.Renew(*cert, true, false, "")
	if err != nil {
		return
	}

	if err = c.certRepo.Save(cert); err != nil {
		return
	}

	return
}
