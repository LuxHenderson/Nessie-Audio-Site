# Nessie Audio — Custom E-Commerce and Portfolio Platform for a Small Audio Business

## Project Overview

Nessie Audio is a full-stack website I designed and built to serve as the central business hub for a small audio production operation. The site consolidates a music portfolio, merchandise storefront, booking system, and contact pipeline into a single owned platform. It replaces a prior workflow that depended on third-party freelance marketplaces and fragmented communication channels.

## Problem Statement

Before this project, business operations were spread across platforms like Fiverr and Upwork, each imposing its own fee structures, branding constraints, and limitations on how services could be presented. There was no single destination where a visitor could browse the portfolio, purchase merchandise, and initiate a booking in one session. I needed an owned, flexible solution that could unify these functions under one roof without ongoing platform commissions or design restrictions.

## Goals & Requirements

The site needed to support several core functions: displaying a music portfolio, selling print-on-demand merchandise with variant selection, processing payments, and accepting booking inquiries through a validated contact form. Non-functional requirements included mobile responsiveness, accessibility compliance, search engine visibility, and the ability to deploy and operate the site with minimal ongoing infrastructure cost. I also required environment-specific configuration so development, staging, and production could each run with isolated settings.

## Design Decisions

I structured the site around a persistent navigation header with a fixed atmospheric background and a Three.js particle fog effect to establish visual identity without relying on heavy imagery. Information architecture follows a flat hierarchy: portfolio, merch, gallery, tour dates, about, and contact are all one click from any page. I implemented a dark mode toggle with localStorage persistence and chose a mobile-first responsive layout using CSS Grid and Flexbox, with breakpoints at 600px and 900px. Accessibility was a priority from the start, so I used semantic HTML, ARIA landmarks, skip-to-content links, screen reader announcements for cart updates, and keyboard-navigable lightbox and menu components.

## Technical Stack & Implementation

The frontend is built with vanilla HTML, CSS, and JavaScript without a framework, which keeps the bundle size minimal and eliminates build tooling dependencies. The backend is a Go server using gorilla/mux for routing and SQLite for persistence, deployed as a multi-stage Docker image on Railway. Payments are handled through Stripe Checkout, and merchandise fulfillment is automated via the Printful API, creating a hands-off order pipeline from purchase to shipment. Email notifications use SMTP through Gmail, and the booking form submits through Formspree. Environment detection is automatic, reading Railway environment variables, hostname patterns, or marker files to select the correct configuration without manual flags.

## Key Challenges & Solutions

Integrating two external APIs — Stripe for payments and Printful for fulfillment — introduced reliability concerns, since either service could experience downtime. I implemented a circuit breaker pattern for both clients, configured to open after five consecutive failures and reset after sixty seconds, preventing cascading request storms during outages. Webhook processing required idempotency guarantees, which I addressed by storing Stripe event IDs with a unique constraint and checking for duplicates before processing. Rate limiting was another challenge; I built a token bucket system with per-IP tracking and endpoint-specific configurations, allowing higher throughput for product browsing while restricting checkout attempts to prevent abuse. On the frontend, Chrome's tab throttling would freeze the Three.js fog animation, so I added a watchdog timer and visibility change listeners to detect stalls and restart the render loop.

## Outcome

The finished site consolidates portfolio presentation, merchandise sales, booking intake, and business information into a single platform with no per-transaction platform fees beyond standard payment processing costs. The Stripe-to-Printful pipeline automates order fulfillment end-to-end, removing manual steps that previously required coordinating across separate services. The site scores well on accessibility audits, serves proper Open Graph and Twitter Card metadata for social sharing, and includes a web app manifest for installability. Scheduled database backups run daily with thirty-day retention, and a health check endpoint enables Railway to monitor uptime and auto-restart on failure.

## Learnings & Next Steps

Building the backend in Go reinforced the value of explicit error handling and dependency injection, particularly when coordinating multiple external services with different failure modes. Choosing SQLite simplified deployment significantly — the entire application is a single binary plus a database file — but it will require migration to PostgreSQL if concurrent write volume increases. Planned improvements include adding a digital products storefront, implementing server-side rendered product pages for faster initial loads, and expanding the inventory tracking system to support alerting thresholds with scheduled checks rather than manual triggers.
