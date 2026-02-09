# Nessie Audio

A full-stack website serving as the central business hub for Nessie Audio, a small audio production operation. The site consolidates a music portfolio, merchandise storefront, booking system, and contact pipeline into a single owned platform.

## Project Purpose

This project replaces a prior workflow that relied on third-party freelance marketplaces (Fiverr, Upwork) and fragmented communication channels. The goal was to eliminate per-transaction platform fees, remove branding constraints imposed by those marketplaces, and provide a single destination where visitors can browse the portfolio, purchase merchandise, and submit booking inquiries without leaving the site.

## Project Status

v1.0 is deployed and live in production on Railway. The core feature set — portfolio, merch store with Stripe checkout, Printful fulfillment, and booking form — is complete and operational. The project is actively maintained and evolving. A digital products storefront is planned but not yet implemented (the `nessie-digital.html` page is a placeholder).

## Tech Stack

| Layer        | Technology                                                                 |
|--------------|----------------------------------------------------------------------------|
| Frontend     | Vanilla HTML, CSS, JavaScript (no framework, no build step)                |
| Backend      | Go 1.23, gorilla/mux router                                                |
| Database     | SQLite (single-file, embedded via go-sqlite3 with CGO)                     |
| Payments     | Stripe Checkout (server-side session creation, webhook verification)       |
| Fulfillment  | Printful API (automated print-on-demand order submission)                  |
| Email        | SMTP via Gmail (order confirmations, low-stock alerts, error notifications)|
| Booking Form | Formspree (third-party form submission endpoint)                           |
| Deployment   | Docker (multi-stage build), Railway (container hosting with health checks) |
| Visual FX    | Three.js (WebGL particle fog effect)                                       |

**Why vanilla frontend:** A framework was unnecessary for a site with static pages and a small amount of dynamic content (merch store, cart). Vanilla JS keeps the bundle size minimal, eliminates build tooling, and avoids framework churn for a project maintained by one developer.

**Why SQLite:** The application handles low to moderate write volume. SQLite eliminates the need for a separate database server, simplifies deployment to a single binary plus a database file, and keeps infrastructure costs at zero beyond the Railway container.

## Architecture Overview

The Go backend serves both the API and the frontend static files from a single process. In production, the Docker image packages the compiled binary alongside all HTML, CSS, JS, and image assets.

```
Frontend (static files)
    |
    v
Go Server (gorilla/mux)
    |
    |-- Middleware stack (in order):
    |     Recovery -> RequestID -> HTTPS Redirect -> Security Headers -> Logging -> CORS -> Rate Limit
    |
    |-- /api/v1/products         GET     Product catalog
    |-- /api/v1/products/{id}    GET     Product detail with variants
    |-- /api/v1/orders           POST    Create order
    |-- /api/v1/orders/{id}      GET     Retrieve order
    |-- /api/v1/cart/checkout    POST    Stripe session from cart
    |-- /api/v1/inventory/*      GET/PUT Inventory management
    |-- /api/v1/config           GET     Public Stripe key
    |-- /webhooks/stripe         POST    Stripe event processing
    |-- /webhooks/printful/{t}   POST    Printful event processing
    |-- /health                  GET     Component health status
    |-- /sitemap.xml             GET     SEO sitemap
    |-- /*                       GET     Static file serving (catch-all)
    |
    |-- SQLite (nessie_store.db)
    |
    |-- External services:
          Stripe  (circuit breaker: 5 failures / 60s reset)
          Printful (circuit breaker: 5 failures / 60s reset)
          SMTP
```

The order flow: cart checkout creates a Stripe session, Stripe's webhook confirms payment, the backend submits the order to Printful for fulfillment, and Printful's webhook updates tracking information. Each external service call is wrapped in a circuit breaker to prevent cascading failures during outages.

## Key Features

- **Merchandise store** with product variants (size, color, capacity), dynamic pricing, and add-to-cart with localStorage persistence
- **Stripe Checkout** integration with server-side session creation and webhook-driven order lifecycle
- **Printful fulfillment** automation: paid orders are submitted to Printful for print-on-demand production and shipping
- **Circuit breaker pattern** on Stripe and Printful clients (5-failure threshold, 60-second reset, half-open probe)
- **Inventory tracking** with configurable per-variant thresholds; print-on-demand items default to unlimited stock
- **Rate limiting** via token bucket algorithm with per-IP tracking and endpoint-specific configurations
- **Security headers** including CSP, HSTS, X-Frame-Options, Permissions-Policy, and Referrer-Policy
- **Webhook idempotency** via unique constraint on Stripe event IDs
- **Request tracing** with UUID per request, included in logs and error responses
- **Structured logging** to file and stdout with severity levels; critical errors trigger admin email
- **Scheduled database backups** daily at 3:00 AM with gzip compression (30-day daily retention, 12-month monthly retention)
- **Accessibility** features: semantic HTML, ARIA landmarks, skip-to-content links, screen reader announcements, keyboard navigation
- **Dark mode** toggle with localStorage persistence
- **Three.js fog effect** with Chrome tab-throttling workaround (watchdog timer, visibility detection)
- **Form validation** with input sanitization, honeypot spam prevention, and double-submission blocking
- **SEO** with per-page Open Graph and Twitter Card meta tags, robots.txt, and a generated sitemap
- **PWA manifest** for home screen installability
- **Clean URLs** with automatic `.html` extension resolution
- **Graceful shutdown** with 30-second drain timeout

## Local Development

### Prerequisites

- Go 1.23+
- GCC (required for go-sqlite3 CGO compilation)
- A local web server for the frontend (e.g., VS Code Live Server on port 5500)

### Backend Setup

1. Navigate to the Backend directory:
   ```
   cd Backend
   ```

2. Create a `.env` (or `.env.development`) file with the required variables:
   ```
   PORT=8080
   STRIPE_SECRET_KEY=sk_test_...
   STRIPE_PUBLISHABLE_KEY=pk_test_...
   STRIPE_WEBHOOK_SECRET=whsec_...
   PRINTFUL_API_KEY=...
   PRINTFUL_WEBHOOK_SECRET=...
   PRODUCTION_DOMAIN=
   SMTP_HOST=smtp.gmail.com
   SMTP_PORT=587
   SMTP_USERNAME=...
   SMTP_PASSWORD=...
   SMTP_FROM_EMAIL=...
   ADMIN_EMAIL=...
   ```
   The server will start without API keys but Stripe and Printful integrations will be non-functional. Environment detection defaults to `development` on local machines.

3. Run the server:
   ```
   go run ./cmd/server
   ```
   The server starts on `http://localhost:8080`. It will automatically run migrations, seed the products/variants tables if empty, and create an initial backup.

### Frontend

Open the HTML files via a local server (e.g., Live Server on port 5500). The frontend's `config.js` detects localhost and routes API calls to `http://localhost:8080/api/v1`. When accessing from a LAN IP (e.g., 192.168.x.x), `resolveAssetUrl()` rewrites image URLs to match the current host.

### Stripe Webhooks (Local)

To test the full checkout flow locally, use the Stripe CLI to forward events:
```
stripe listen --forward-to localhost:8080/webhooks/stripe
```

## Content & Data Management

### Products and Variants

Products and variants are seeded on first startup from hardcoded data in `Backend/cmd/server/main.go` (the `seedProductsIfEmpty` and `seedVariantsIfEmpty` functions). Each product maps to a Printful sync product via `printful_id`, and each variant maps to a Printful sync variant via `printful_variant_id`. To add or modify products, update the seed data and either wipe the database or insert directly via SQL.

### Database

SQLite database file: `nessie_store.db` (created automatically on first run). Schema is managed through migrations in `Backend/internal/migrations/` and incremental `ALTER TABLE` statements in `Backend/internal/database/db.go`. In production on Railway, the database is stored on a persistent volume at the path specified by `RAILWAY_VOLUME_MOUNT_PATH`.

### Static Assets

Product images live in `Product Photos/` at the project root, organized by product name. The background image is `Nessie Audio 2026.jpg`. Music files are in `Music/`. All static assets are copied into the Docker image at build time under `/app/static/`.

### Backups

The backup system runs automatically. Daily backups are gzip-compressed and stored in `backups/daily/` (retained 30 days). Monthly backups go to `backups/monthly/` (retained 12 months). An initial backup is created on each server startup.

## Known Constraints / Tradeoffs

- **SQLite concurrency:** SQLite supports only one writer at a time. This is acceptable at current traffic levels but would require migration to PostgreSQL if concurrent write volume increases significantly.
- **Seed data in application code:** Products are hardcoded in `main.go` rather than managed through an admin interface or external CMS. This is a deliberate simplification — product changes are infrequent and a full admin panel would be premature.
- **No server-side rendering:** Product detail pages fetch data client-side, which means the initial HTML served to crawlers does not contain product-specific content. Meta tags are updated dynamically via JavaScript, which most modern crawlers handle but is not as reliable as SSR for SEO.
- **Single-process architecture:** The backend serves both the API and static files from one process. This simplifies deployment but means a backend restart briefly interrupts static file serving.
- **Formspree dependency:** The booking form submits through Formspree rather than the Go backend. This was a practical choice to avoid implementing email delivery and spam filtering for form submissions, but it introduces a third-party dependency for a core workflow.

## Future Improvements

- Digital products storefront (currently a placeholder page)
- Server-side rendering or pre-rendering for product pages to improve SEO reliability
- Admin interface for product and order management
- Migration path from SQLite to PostgreSQL for higher concurrency
- Scheduled low-stock alert checks (currently manual trigger only)
- Automated test suite for backend handlers and integration flows

## Notes

- The project was originally scaffolded under the name "Naevermore" and has since been renamed to Nessie Audio. Some internal references (e.g., the localStorage key `naevermore-theme` for dark mode) still reflect the original name.
- Environment detection is automatic and does not require a manual flag. The priority order is: `ENV` variable, `RAILWAY_ENVIRONMENT`, hostname pattern matching, marker files (`.production`, `.staging`), then defaults to `development`.
- The frontend uses no build tooling. All JavaScript files are loaded directly via `<script>` tags. `config.js` is the shared configuration module that all other scripts depend on for API endpoint resolution.
- Cache-Control headers are set per file type: HTML files use `no-cache` to prevent CDN/proxy stale pages; static assets are cached for one hour.
- The circuit breaker implementation uses three states (closed, open, half-open) and is applied independently to Stripe and Printful clients. When open, requests fail immediately with `ErrCircuitOpen` rather than waiting for a timeout.
