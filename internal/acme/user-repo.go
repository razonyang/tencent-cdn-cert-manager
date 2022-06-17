package acme

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"path"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/registration"
	"github.com/razonyang/tencent-cdn-cert-manager/internal/helper"
	"github.com/sirupsen/logrus"
)

const (
	dirPerm  = 0o700
	filePerm = 0o600
)

type UserRepository interface {
	Get(email string) (registration.User, error)
	Save(user registration.User) error
}

type FilesystemUserRepository struct {
	root           string
	caServer       string
	caServerHost   string
	accountFile    string
	privateKeyFile string
}

func NewFilesystemUserRepository(root, caServer string) *FilesystemUserRepository {
	return &FilesystemUserRepository{
		root:           path.Join(root, "accounts"),
		caServer:       caServer,
		accountFile:    "account.json",
		privateKeyFile: "private.key",
	}
}

func (fr *FilesystemUserRepository) getCaServerHost() (string, error) {
	if fr.caServerHost == "" {
		host, err := FormatCAServerHost(fr.caServer)
		if err != nil {
			return "", err
		}

		fr.caServerHost = host
	}

	return fr.caServerHost, nil
}

func (fr *FilesystemUserRepository) getAccountFilename(email string) (filename string, err error) {
	host, err := fr.getCaServerHost()
	if err != nil {
		return
	}

	return path.Join(fr.root, host, email, fr.accountFile), nil
}

func (fr *FilesystemUserRepository) Get(email string) (user registration.User, err error) {
	accountFile, err := fr.getAccountFilename(email)
	if err != nil {
		return
	}
	data, err := os.ReadFile(accountFile)
	if err == nil {
		user := &User{}
		if err := json.Unmarshal(data, user); err != nil {
			logrus.Errorf("unable to unmarshal account %q: %s", email, err)
		} else {
			keyFilename := path.Join(path.Dir(accountFile), "private.key")
			keyFile, err := os.ReadFile(keyFilename)
			if err == nil {
				keyBlock, _ := pem.Decode(keyFile)
				var key crypto.PrivateKey
				switch keyBlock.Type {
				case "RSA PRIVATE KEY":
					key, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
				case "EC PRIVATE KEY":
					key, err = x509.ParseECPrivateKey(keyBlock.Bytes)
				default:
					err = fmt.Errorf("unsupported private key type: %s", keyBlock.Type)
				}
				if err == nil {
					user.Key = key
					return user, nil
				} else {
					logrus.Errorf("unable to parse private key: %s", err)
				}
			}
		}
	} else if os.IsNotExist(err) {
		logrus.Infof("account %q does not exist", email)
	}

	return fr.create(email)
}

func (fr *FilesystemUserRepository) create(email string) (registration.User, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("unable to generate private key for account %q: %s", email, err)
	}

	return &User{
		email: email,
		Key:   privateKey,
	}, nil
}

func (fr *FilesystemUserRepository) Save(user registration.User) error {
	accountFile, err := fr.getAccountFilename(user.GetEmail())
	if err != nil {
		return err
	}
	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("unable to marshal account %q: %s", user.GetEmail(), err)
	}

	if err := helper.CreatePath(path.Dir(accountFile), dirPerm); err != nil {
		return fmt.Errorf("unable to create folder %q: %s", path.Dir(accountFile), err)
	}

	if err := os.WriteFile(accountFile, data, filePerm); err != nil {
		return fmt.Errorf("unable to save account %q: %s", user.GetEmail(), err)
	}
	keyFilename := path.Join(path.Dir(accountFile), "private.key")
	keyFile, err := os.Create(keyFilename)
	if err != nil {
		return fmt.Errorf("unable to create private key: %s", err)
	}
	err = pem.Encode(keyFile, certcrypto.PEMBlock(user.GetPrivateKey()))
	if err != nil {
		return fmt.Errorf("unable to write private key: %s", err)
	}

	return nil
}
