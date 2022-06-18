# Tencent CDN Cert Manager

Manage your Tencent cloud CDN certificates, including obtaining and renewing certificates automatically.

## Environments Variables

The program will load variables from `.env` of current working directory.

> You can find all variables on [.env.example](.env.example).

| Name | Default | Description | References |
|---|---|---|---|
| `TENCENT_REGION` | - | Tencent Region | https://pkg.go.dev/github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions
| `TENCENT_API_SECRET_ID` | - | Tencent API secret Id |
| `TENCENT_API_SECRET_KEY` | - | Tencent API secret key |
| `CERT_MANAGER_DOMAINS` | - | Tencent CDN domains, separated by commas |
| `CERT_MANAGER_EMAIL` | - | Your email address|
| `CA_DIR_URL` | `https://acme-staging-v02.api.letsencrypt.org/directory` | CA directory URL | Replace it with `https://acme-v02.api.letsencrypt.org/directory` in production
| `CERT_MANAGER_INTERVAL` | `@hourly` | The cron job interval | https://pkg.go.dev/github.com/robfig/cron#hdr-Usage
| `CERT_MANAGER_DAYS` | `30` | Renew certificates that expires within `n` days |
| `DNS_PROVIDER` | - | `cloudlfare`, `alidns` or `tencentcloud` | DNS provider |
| **DNS PROVIDER VARIABLES** | - | | [`cloudlfare`](https://go-acme.github.io/lego/dns/cloudflare/), [`alidns`](https://go-acme.github.io/lego/dns/alidns/) or [`tencentcloud`](https://go-acme.github.io/lego/dns/tencentcloud/)

## Testing

I recommend testing your `.env` with `CA_DIR_URL = https://acme-staging-v02.api.letsencrypt.org/directory`, Otherwise, you may encounter a rate limit problem.

And change `CA_DIR_URL` as `https://acme-v02.api.letsencrypt.org/directory` if your config is OK.

## Docker

```bash
$ docker run \
  -v "$PWD/.env:/app/.env:ro" \
  -v data:/data \
  --name tccm \
  razonyang/tencent-cdn-cert-manager
```

- `/data` stores user private keys and SSL certificates that used to renew certificates.

> You can also specify variables via `-e` instead of mounting `.env` file.
