# Security and Secrets Management

## Important: Never Commit Real Secrets

This project uses environment variables for sensitive configuration. Follow these guidelines:

### Local Development

1. **Copy the template:**
   ```bash
   cp .env.example .env
   ```

2. **Add your real credentials to `.env`:**
   - Printful API Key
   - Stripe Secret Key
   - Stripe Publishable Key
   - Stripe Webhook Secret

3. **Never commit `.env` file** - it's in `.gitignore` for security

### Files Tracked by Git

- ✅ `.env.example` - Contains placeholder values only
- ✅ `README.md` - Contains placeholder examples only
- ❌ `.env` - NEVER commit (contains real secrets)
- ❌ `.env.local` - NEVER commit
- ❌ `.env.production` - NEVER commit

### If You Accidentally Commit Secrets

1. **Immediately rotate the exposed credentials:**
   - Generate new Printful API key
   - Generate new Stripe keys
   - Update your local `.env` file

2. **Remove from Git history** (if needed):
   ```bash
   git filter-branch --force --index-filter \
     "git rm --cached --ignore-unmatch Backend/.env" \
     --prune-empty --tag-name-filter cat -- --all
   ```

3. **Force push** (use with caution):
   ```bash
   git push origin --force --all
   ```

### Deployment

For production deployments:
- Use environment variable management in your hosting platform (Heroku, Railway, AWS, etc.)
- Never hardcode secrets in source code
- Use different API keys for development vs production
- Enable Stripe test mode for development

## Current Security Status

✅ Real secrets removed from `.env.example`  
✅ `.gitignore` configured to exclude all `.env*` files  
✅ README documents proper credential management  
✅ Copilot instructions include security warning  
✅ Repository ready for safe push to GitHub
