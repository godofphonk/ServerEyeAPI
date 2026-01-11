# Production Secrets Setup

## Required GitHub Secrets

Add these secrets to your GitHub repository Settings > Secrets and variables > Actions:

### Database Secrets
- `POSTGRES_PASSWORD` - PostgreSQL password (min 12 chars)
- `DATABASE_URL` - Full database connection string

### Application Secrets  
- `JWT_SECRET` - JWT signing key (min 32 chars)
- `WEBHOOK_SECRET` - Webhook verification secret (min 32 chars)

### Infrastructure Secrets
- `PROD_HOST` - Production server hostname/IP
- `PROD_USER` - SSH username for deployment
- `PROD_SSH_KEY` - SSH private key for deployment

### Configuration Variables
Add these to Repository Variables (not secrets):
- `WEB_URL` - Production web URL (e.g., https://api.servereye.dev)

## Example Values
```bash
# Generate secure secrets:
openssl rand -base64 32  # JWT_SECRET
openssl rand -base64 32  # WEBHOOK_SECRET  
openssl rand -base64 16  # POSTGRES_PASSWORD
```

## Security Notes
- Never commit secrets to repository
- Use GitHub Secrets for sensitive data
- Use Variables for non-sensitive config
- Rotate secrets regularly
- Monitor access logs
