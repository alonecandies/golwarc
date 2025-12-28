# Security Policy

## Supported Versions

We release patches for security vulnerabilities for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take the security of Golwarc seriously. If you believe you have found a security vulnerability, please report it to us as described below.

### Where to Report

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them via email to: security@golwarc.dev

You should receive a response within 48 hours. If for some reason you do not, please follow up via email to ensure we received your original message.

### What to Include

Please include the following information in your report:

- Type of vulnerability (e.g., SQL injection, XSS, SSRF, etc.)
- Full paths of source file(s) related to the manifestation of the issue
- The location of the affected source code (tag/branch/commit or direct URL)
- Any special configuration required to reproduce the issue
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the issue, including how an attacker might exploit it

### What to Expect

- We will acknowledge receipt of your vulnerability report
- We will confirm the vulnerability and determine its impact
- We will release a fix as soon as possible, depending on complexity
- We will credit you in the security advisory (unless you prefer to remain anonymous)

## Security Best Practices

### Production Deployment

When deploying Golwarc in production, follow these security best practices:

#### 1. Enable TLS for All Connections

```yaml
cache:
  redis:
    addr: "redis.example.com:6379"
    tls:
      enabled: true
      ca_cert: "/path/to/ca.crt"
      verify: true

database:
  mysql:
    host: "mysql.example.com"
    port: 3306
    tls:
      enabled: true
      ca_cert: "/path/to/ca.crt"
```

#### 2. Use Strong Authentication

- Never use default passwords
- Use strong, unique passwords for all database connections
- Rotate credentials regularly
- Store credentials securely (environment variables, secrets manager)

#### 3. Configure Rate Limiting

```yaml
crawlers:
  colly:
    rate_limit:
      delay: 2s
      random_delay: 1s
      max_concurrent: 5
```

#### 4. Validate All Input

- All crawler URLs are validated for SSRF protection
- Never trust user input
- Sanitize data before storage

#### 5. Network Security

- Use firewalls to restrict access to databases and caches
- Deploy services in private networks
- Use VPN or bastion hosts for administrative access

#### 6. Keep Dependencies Updated

- Enable Dependabot for automated dependency updates
- Regularly run `govulncheck ./...` to scan for vulnerabilities
- Review and test updates before deploying

#### 7. Monitoring and Logging

- Enable structured logging
- Monitor for unusual patterns
- Set up alerts for security events
- Regularly review logs

### Configuration Security

#### Never Commit Secrets

```bash
# Use .gitignore for sensitive files
config.yaml
.env
*.key
*.crt
*.pem
```

#### Use Environment Variables

```bash
export GOLWARC_DATABASE_MYSQL_PASSWORD="secure-password"
export GOLWARC_CACHE_REDIS_PASSWORD="redis-password"
```

#### Secure File Permissions

```bash
chmod 600 config.yaml
chmod 600 .env
```

## Known Security Considerations

### URL Validation

Golwarc includes SSRF protection that blocks:

- `file://` and `javascript://` schemes
- Localhost and loopback addresses
- Private IP ranges (10.x.x.x, 192.168.x.x, 172.16-31.x.x)

However, always validate and sanitize URLs from untrusted sources.

### Rate Limiting

Configure appropriate rate limits to prevent:

- Denial of service attacks on crawl targets
- Resource exhaustion on your infrastructure
- IP blocking by target websites

### Database Queries

While we use GORM which provides SQL injection protection, always:

- Use parameterized queries
- Validate input data
- Apply principle of least privilege for database users

## Security Updates

We will notify users of security updates through:

1. GitHub Security Advisories
2. Release notes in CHANGELOG.md
3. Git tags for security releases

Security releases will be tagged with the format: `vX.Y.Z-security`

## Vulnerability Disclosure Timeline

- **Day 0**: Vulnerability reported
- **Day 1-2**: Acknowledge receipt and begin investigation
- **Day 3-7**: Confirm vulnerability and develop fix
- **Day 8-14**: Test and release patch
- **Day 15**: Public disclosure (if applicable)

## Dependencies

We regularly audit our dependencies for security vulnerabilities:

```bash
# Check for vulnerable dependencies
govulncheck ./...

# Update dependencies
go get -u ./...
go mod tidy
```

## Contact

For security concerns, contact: security@golwarc.dev

For general questions, use [GitHub Discussions](https://github.com/alonecandies/golwarc/discussions)

---

**Last Updated:** December 28, 2025
