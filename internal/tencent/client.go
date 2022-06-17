package tencent

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/razonyang/tencent-cdn-cert-manager/internal/helper"
	"github.com/sirupsen/logrus"
	cdn "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn/v20180606"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

var (
	certFrom    = "Let's Encrypt"
	certMessage = "TENCENT CDN CERT MANAGER"
)

type Client struct {
	*cdn.Client
}

func NewClientEnv() (*Client, error) {
	return NewClient(os.Getenv("TENCENT_REGION"), os.Getenv("TENCENT_API_SECRET_ID"), os.Getenv("TENCENT_API_SECRET_KEY"))
}

func NewClient(region, apiSecretId, APISecretKey string) (*Client, error) {
	credential := common.NewCredential(apiSecretId, APISecretKey)
	client, err := cdn.NewClient(credential, region, profile.NewClientProfile())
	if err != nil {
		return nil, err
	}
	return &Client{
		client,
	}, nil
}

func (c *Client) fetch(domain string) (detail *cdn.DetailDomain, err error) {
	request := cdn.NewDescribeDomainsConfigRequest()
	filterName := "domain"
	filter := &cdn.DomainFilter{
		Name:  &filterName,
		Value: []*string{&domain},
	}
	request.Filters = append(request.Filters, filter)
	response, err := c.DescribeDomainsConfig(request)
	if err != nil {
		return
	}

	for _, detailDomain := range response.Response.Domains {
		if *detailDomain.Domain == domain {
			detail = detailDomain
			break
		}
	}
	if detail == nil {
		err = fmt.Errorf("[%s] does not exist", domain)
		return
	}

	return
}

func (c *Client) ValidateCertificate(domain string) (notSet bool, renew bool, err error) {
	detail, err := c.fetch(domain)
	if err != nil {
		logrus.Infof("[%s] failed to fetch: %s\n", domain, err)
		return
	}

	if detail.Https == nil || detail.Https.CertInfo == nil {
		notSet = true
		logrus.Infof("[%s] certificate doesn't exits\n", domain)
		return
	}

	// Check if the&& os.IsNotExist(err)certificate expire time.
	if detail.Https.CertInfo.ExpireTime == nil {
		notSet = true
		logrus.Infof("[%s] certificate expire time unknown\n", domain)
		return
	}
	expireTime, err := time.Parse("2006-01-02 15:04:05", *detail.Https.CertInfo.ExpireTime)
	if err != nil {
		logrus.Error("[%s] certificate expire time parsed fails\n", domain)
		return
	}
	logrus.Infof("[%s] certificate expires within: %s", domain, expireTime.Sub(time.Now()))
	daysEnv := helper.Getenv("CERT_MANAGER_DAYS", "30")
	days, err := strconv.Atoi(daysEnv)
	if err != nil {
		err = fmt.Errorf("[%s] invalid CERT_MANAGER_DAYS: %s", domain, daysEnv)
		return
	}
	// Renew the certificate if expired or expires in n days.
	if time.Until(expireTime) < time.Hour*24*time.Duration(days) {
		renew = true
		logrus.Infof("[%s] certificate expired or expires within %d days", domain, days)
		return
	}

	return
}

func (c *Client) UploadCertificate(domain string, cert *certificate.Resource) (err error) {
	logrus.Infof("[%s] certificate uploading...\n", domain)
	request := cdn.NewUpdateDomainConfigRequest()
	certificate := string(cert.Certificate)
	privateKey := string(cert.PrivateKey)
	request.Domain = &domain
	https := "on"
	p, _ := pem.Decode(cert.Certificate)
	x509Cert, err := x509.ParseCertificate(p.Bytes)
	if err != nil {
		return err
	}
	expiredTime := x509Cert.NotAfter.Format("2006-01-02 15:04:05")
	request.Https = &cdn.Https{
		Switch: &https,
		CertInfo: &cdn.ServerCert{
			Certificate: &certificate,
			PrivateKey:  &privateKey,
			ExpireTime:  &expiredTime,
			Message:     &certMessage,
			From:        &certFrom,
		},
	}

	_, err = c.UpdateDomainConfig(request)
	if err != nil {
		return err
	}

	logrus.Infof("[%s] certificate uploaded\n", domain)
	return nil
}
