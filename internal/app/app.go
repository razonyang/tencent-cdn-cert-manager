package app

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/razonyang/tencent-cdn-cert-manager/internal/acme"
	"github.com/razonyang/tencent-cdn-cert-manager/internal/helper"
	"github.com/razonyang/tencent-cdn-cert-manager/internal/tencent"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

const (
	ENV_PRODUCATION = "producation"
	ENV_DEVELOPMENT = "development"
)

type Application struct {
	Email   string
	Domains []string
	client  *tencent.Client
}

func New() *Application {
	app := &Application{
		Email: os.Getenv("CERT_MANAGER_EMAIL"),
	}
	domains := strings.TrimSpace(os.Getenv("CERT_MANAGER_DOMAINS"))
	if domains != "" {
		app.Domains = strings.Split(domains, ",")
	}
	return app
}

func (app *Application) Run() error {
	if app.Email == "" {
		return fmt.Errorf("please specify email via the CERT_MANAGER_EMAIL env var")
	}
	if len(app.Domains) == 0 {
		return errors.New("please specify domains via the CERT_MANAGER_DOMAINS env var")
	}

	client, err := tencent.NewClientEnv()
	if err != nil {
		return err
	}
	app.client = client

	c := cron.New()
	interval := strings.TrimSpace(helper.Getenv("CERT_MANAGER_INTERVAL", "@hourly"))
	for _, domain := range app.Domains {
		logrus.Infof("[%s] monitor interval: %s\n", domain, interval)
		if _, err := c.AddFunc(interval, func() { app.monitor(domain) }); err != nil {
			logrus.Fatalf("[%s] unable to monitor with interval %q: %s", domain, interval, err)
		}
		// run the job immediately.
		go app.monitor(domain)
	}
	c.Run()

	return nil
}

func (app *Application) monitor(domain string) {
	logrus.Infof("[%s] job started", domain)
	defer func() {
		if err := recover(); err != nil {
			if _, ok := err.(*logrus.Entry); !ok {
				logrus.Errorf("[%s] job failed: %s\n", domain, err)
			} else {
				logrus.Errorf("[%s] job failed\n", domain)
			}
		} else {
			logrus.Infof("[%s] job finished\n", domain)
		}
	}()

	notSet, expired, err := app.client.ValidateCertificate(domain)
	if err != nil {
		logrus.Panic(err)
	}
	if notSet {
		app.createOrRenew(domain, false)
	}
	if expired {
		app.createOrRenew(domain, true)
	}
}

func (app *Application) createOrRenew(domain string, renew bool) {
	acmeClient, err := acme.NewClient(app.Email, os.Getenv("DNS_PROVIDER"))
	if err != nil {
		logrus.Panicf("[%s] unable to create the ACME client: %s", domain, err)
	}
	var cert *certificate.Resource
	if renew {
		cert, err = acmeClient.RenewCertificate(domain)
		if err != nil {
			logrus.Errorf("[%s] unable to renew certificate, creating a new certificate: %s", domain, err)
		}
	}
	if !renew || (renew && err != nil) {
		cert, err = acmeClient.ObtainCertificate(domain)
	}
	if err != nil {
		logrus.Panicf("[%s] unable to obtain certificate: %s", domain, err)
	}

	if err := app.client.UploadCertificate(domain, cert); err != nil {
		logrus.Panicf("[%s] unable to upload the CDN certificate: %s", domain, err)
	}
}
