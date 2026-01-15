# Helm Chart Configuration

This guide details the configuration options for the `readwillbe` Helm chart.

## Prerequisites

- Kubernetes cluster
- Helm 3+ installed

## Installation

```bash
helm install my-readwillbe ./charts/readwillbe
```

## Configuration

The following table lists the configurable parameters of the ReadWillBe chart and their default values.

### General Settings

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `1` |
| `image.repository` | Image repository | `readwillbe` |
| `image.tag` | Image tag | `latest` |
| `service.port` | Kubernetes Service port | `80` |
| `ingress.enabled` | Enable Ingress | `false` |
| `ingress.hosts` | List of ingress hosts | `chart-example.local` |

### Application Configuration (`env`)

These values map directly to environment variables in the container.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `env.READWILLBE_PORT` | Application listening port | `":8080"` |
| `env.READWILLBE_LOG_LEVEL` | Logging level | `"info"` |
| `env.READWILLBE_ALLOW_SIGNUP` | Enable user registration | `"true"` |
| `env.READWILLBE_HOSTNAME` | Public URL of the app | `http://localhost:8080` |
| `env.TZ` | Container Timezone | `"America/New_York"` |

### Persistence

| Parameter | Description | Default |
|-----------|-------------|---------|
| `persistence.enabled` | Enable data persistence | `true` |
| `persistence.size` | PVC Size | `1Gi` |
| `persistence.path` | Mount path for data | `/data` |

### Secrets

**Important:** For production, it is recommended to pass secrets via a separate values file or `--set` arguments rather than committing them to `values.yaml`.

| Parameter | Description | Required |
|-----------|-------------|:--------:|
| `secrets.cookieSecret` | Session encryption key (min 32 chars) | **Yes** |
| `secrets.vapidPublicKey` | VAPID Public Key for Push Notifications | No |
| `secrets.vapidPrivateKey` | VAPID Private Key for Push Notifications | No |
| `secrets.smtpPassword` | SMTP Password (if using SMTP) | No |
| `secrets.resendApiKey` | Resend API Key (if using Resend) | No |

### Email Configuration

Configure one of the following providers under the `email` section.

**SMTP (`email.provider: "smtp"`)**
- `email.smtp.host`
- `email.smtp.port`
- `email.smtp.username`
- `email.smtp.from`
- `email.smtp.tls`

**Resend (`email.provider: "resend"`)**
- `email.resend.from`

## Example `values-prod.yaml`

```yaml
ingress:
  enabled: true
  hosts:
    - host: read.my-domain.com
      paths:
        - path: /
          pathType: Prefix

secrets:
  cookieSecret: "CHANGE_ME_TO_A_SECURE_RANDOM_STRING"
  smtpPassword: "my-smtp-password"

email:
  provider: "smtp"
  smtp:
    host: "smtp.gmail.com"
    username: "me@gmail.com"
    from: "ReadWillBe <me@gmail.com>"
```

Install with custom values:

```bash
helm install readwillbe ./charts/readwillbe -f values-prod.yaml
```
