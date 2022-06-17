package acme

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/razonyang/tencent-cdn-cert-manager/internal/helper"
)

type CertRepository interface {
	Get(domain string) (*certificate.Resource, error)
	Save(*certificate.Resource) error
}

type FilesystemCertRepository struct {
	root         string
	caServer     string
	caServerHost string
}

func NewFilesystemCertRepository(root, caServer string) *FilesystemCertRepository {
	return &FilesystemCertRepository{
		root:     path.Join(root, "certificates"),
		caServer: caServer,
	}
}

func (fr *FilesystemCertRepository) getCaServerHost() (string, error) {
	if fr.caServerHost == "" {
		host, err := FormatCAServerHost(fr.caServer)
		if err != nil {
			return "", err
		}

		fr.caServerHost = host
	}

	return fr.caServerHost, nil
}

func (fr *FilesystemCertRepository) Get(domain string) (*certificate.Resource, error) {
	certFile, err := fr.getCertFile(domain)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(certFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read certificate: %s", err)
	}
	var cert certificate.Resource
	if err = json.Unmarshal(data, &cert); err != nil {
		return nil, fmt.Errorf("unable to unmarchal certificate: %s", err)
	}
	dir := path.Dir(certFile)
	if cert.PrivateKey, err = os.ReadFile(path.Join(dir, "private.key")); err != nil {
		return nil, fmt.Errorf("[%s] unable to read private key: %s", domain, err)
	}
	if cert.CSR, err = os.ReadFile(path.Join(dir, "csr.pem")); err != nil {
		return nil, fmt.Errorf("[%s] unable to read CSR: %s", domain, err)
	}
	if cert.Certificate, err = os.ReadFile(path.Join(dir, "cert.pem")); err != nil {
		return nil, fmt.Errorf("[%s] unable to read certificate: %s", domain, err)
	}
	if cert.IssuerCertificate, err = os.ReadFile(path.Join(dir, "issuer-cert.pem")); err != nil {
		return nil, fmt.Errorf("[%s] unable to read issuer certificate: %s", domain, err)
	}

	return &cert, nil
}

func (fr *FilesystemCertRepository) Save(cert *certificate.Resource) error {
	certFile, err := fr.getCertFile(cert.Domain)
	if err != nil {
		return err
	}

	dir := path.Dir(certFile)
	if err = helper.CreatePath(dir, dirPerm); err != nil {
		return err
	}

	data, err := json.Marshal(*cert)
	if err != nil {
		return fmt.Errorf("[%s] unable to marshal certificate: %s", cert.Domain, err)
	}
	if err := os.WriteFile(certFile, data, filePerm); err != nil {
		return fmt.Errorf("[%s] unable to save certificate: %s", cert.Domain, err)
	}
	if err := os.WriteFile(path.Join(dir, "private.key"), cert.PrivateKey, filePerm); err != nil {
		return fmt.Errorf("[%s] unable to save private key: %s", cert.Domain, err)
	}
	if err := os.WriteFile(path.Join(dir, "csr.pem"), cert.CSR, filePerm); err != nil {
		return fmt.Errorf("[%s] unable to save CSR: %s", cert.Domain, err)
	}
	if err := os.WriteFile(path.Join(dir, "cert.pem"), cert.Certificate, filePerm); err != nil {
		return fmt.Errorf("[%s] unable to save certificate: %s", cert.Domain, err)
	}
	if err := os.WriteFile(path.Join(dir, "issuer-cert.pem"), cert.IssuerCertificate, filePerm); err != nil {
		return fmt.Errorf("[%s] unable to save issuer certificate: %s", cert.Domain, err)
	}

	return nil
}

func (fr *FilesystemCertRepository) getCertFile(domain string) (string, error) {
	host, err := fr.getCaServerHost()
	if err != nil {
		return "", err
	}
	return path.Join(fr.root, host, domain, "cert.json"), nil
}

func FormatCAServerHost(server string) (string, error) {
	url, err := url.Parse(server)
	if err != nil {
		return "", fmt.Errorf("unable to parse CA server URL %q: %s", server, err)
	}

	return strings.NewReplacer(":", "_", "/", string(os.PathSeparator)).Replace(url.Host), nil
}
