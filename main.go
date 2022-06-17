package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/razonyang/tencent-cdn-cert-manager/internal/app"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
}

func main() {
	env := ".env"
	if _, err := os.Stat(env); err == nil {
		logrus.Infof("loading env file: %s", env)
		if err := godotenv.Load(env); err != nil {
			logrus.Fatalf("unable to load file: %s", err)
		}
	}

	app := app.New()
	if err := app.Run(); err != nil {
		logrus.Fatal(err)
	}
}
