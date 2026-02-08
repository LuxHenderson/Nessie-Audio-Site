package email

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"strconv"

	"github.com/nessieaudio/ecommerce-backend/internal/config"
	"github.com/nessieaudio/ecommerce-backend/internal/models"
)

// Client handles email sending via SMTP
type Client struct {
	config *config.Config
}

// NewClient creates a new email client
func NewClient(cfg *config.Config) *Client {
	return &Client{
		config: cfg,
	}
}

// OrderConfirmationData holds data for order confirmation emails
type OrderConfirmationData struct {
	OrderID       string
	CustomerName  string
	CustomerEmail string
	Items         []models.OrderItem
	Total         float64
	ShippingInfo  ShippingInfo
}

// ShippingInfo holds shipping details
type ShippingInfo struct {
	Name    string
	Address string
	City    string
	State   string
	Zip     string
	Country string
}

// SendOrderConfirmation sends an order confirmation email
func (c *Client) SendOrderConfirmation(data OrderConfirmationData) error {
	// Generate email HTML
	htmlBody, err := c.generateOrderConfirmationHTML(data)
	if err != nil {
		return fmt.Errorf("failed to generate email HTML: %w", err)
	}

	// Prepare email
	subject := fmt.Sprintf("Order Confirmation - Nessie Audio #%s", data.OrderID)
	to := []string{data.CustomerEmail}

	// Send email
	if err := c.sendEmail(to, subject, htmlBody); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Order confirmation email sent to %s for order %s", data.CustomerEmail, data.OrderID)
	return nil
}

// sendEmail sends an email via SMTP
func (c *Client) sendEmail(to []string, subject, htmlBody string) error {
	// Check if SMTP is configured
	if c.config.SMTPUsername == "" || c.config.SMTPPassword == "" {
		log.Println("WARNING: SMTP not configured, skipping email send")
		return nil
	}

	// SMTP authentication
	auth := smtp.PlainAuth("", c.config.SMTPUsername, c.config.SMTPPassword, c.config.SMTPHost)

	// Build email message
	from := c.config.SMTPFromEmail
	fromName := c.config.SMTPFromName

	// Email headers
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", fromName, from)
	headers["To"] = to[0]
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""

	// Build message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + htmlBody

	// Send email
	addr := fmt.Sprintf("%s:%s", c.config.SMTPHost, c.config.SMTPPort)
	err := smtp.SendMail(addr, auth, from, to, []byte(message))
	if err != nil {
		return fmt.Errorf("smtp error: %w", err)
	}

	return nil
}

// generateOrderConfirmationHTML generates HTML for order confirmation email
func (c *Client) generateOrderConfirmationHTML(data OrderConfirmationData) (string, error) {
	// Inner content template (items loop requires Go templates)
	innerTmpl := `
            <p style="font-size:16px;">Hey {{.CustomerName}},</p>
            <p>Thank you for your order! Your payment has been processed successfully, and we're getting your items ready to ship.</p>

            <div class="info-box">
                <h2>Order Details</h2>
                <div class="detail-row"><span>Order Number:</span> <strong>#{{.OrderID}}</strong></div>
                <div class="detail-row"><span>Email:</span> <strong>{{.CustomerEmail}}</strong></div>
            </div>

            <table class="items-table">
                <thead>
                    <tr>
                        <th>Item</th>
                        <th style="text-align:center;">Qty</th>
                        <th style="text-align:right;">Price</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Items}}
                    <tr>
                        <td>
                            <div class="item-name">{{.ProductName}}</div>
                            {{if .VariantName}}<div class="item-variant">{{.VariantName}}</div>{{end}}
                        </td>
                        <td style="text-align:center;">{{.Quantity}}</td>
                        <td style="text-align:right;">${{printf "%.2f" .TotalPrice}}</td>
                    </tr>
                    {{end}}
                    <tr class="total-row">
                        <td colspan="2">Total</td>
                        <td style="text-align:right;">${{printf "%.2f" .Total}}</td>
                    </tr>
                </tbody>
            </table>

            <div class="info-box">
                <h2>Shipping To</h2>
                <div style="color:#c0c0c0;line-height:1.6;">
                    <strong style="color:#fff;">{{.ShippingInfo.Name}}</strong><br>
                    {{.ShippingInfo.Address}}<br>
                    {{.ShippingInfo.City}}, {{.ShippingInfo.State}} {{.ShippingInfo.Zip}}<br>
                    {{.ShippingInfo.Country}}
                </div>
            </div>

            <div class="note"><strong>What's Next?</strong><br>Your order will be fulfilled by our print-on-demand partner. You'll receive a shipping confirmation email with tracking information once your items are on their way (typically within 2-5 business days).</div>

            <div style="text-align:center;margin:24px 0;"><a href="https://nessieaudio.com/merch" class="cta-button">Continue Shopping</a></div>`

	t, err := template.New("orderConfirmation").Parse(innerTmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return EmailLayout("Order Confirmed!", "&#10003;", buf.String(), false), nil
}

// SendShippingNotification sends a shipping/tracking update email
func (c *Client) SendShippingNotification(customerEmail, orderID, trackingNumber, trackingURL, carrier string) error {
	subject := fmt.Sprintf("Your Nessie Audio Order Has Shipped! #%s", orderID)

	carrierDisplay := carrier
	if carrierDisplay == "" {
		carrierDisplay = "Standard Shipping"
	}

	contentHTML := fmt.Sprintf(`
            <p style="font-size:16px;">Great news! Your Nessie Audio order <strong>#%s</strong> is on its way.</p>
            <div class="tracking-box">
                <p class="tracking-label">Carrier</p>
                <p class="carrier-name">%s</p>
                <p class="tracking-label" style="margin-top:15px;">Tracking Number</p>
                <div class="tracking-number">%s</div>
                <a href="%s" class="cta-button">Track Your Package</a>
            </div>
            %s`,
		orderID, carrierDisplay, trackingNumber, trackingURL,
		NoteBox("<strong>What's Next?</strong><br>Your package is on its way! Tracking information may take up to 24 hours to update after shipment. Click the button above to follow your package in real-time.", false),
	)

	htmlBody := EmailLayout("Your Order Has Shipped!", "&#128230;", contentHTML, false)

	to := []string{customerEmail}
	if err := c.sendEmail(to, subject, htmlBody); err != nil {
		return fmt.Errorf("failed to send shipping notification: %w", err)
	}

	log.Printf("Shipping notification sent to %s for order %s", customerEmail, orderID)
	return nil
}

// SendRawEmail sends a plain text email (for admin alerts)
func (c *Client) SendRawEmail(to, subject, body string) error {
	// Check if SMTP is configured
	if c.config.SMTPUsername == "" || c.config.SMTPPassword == "" {
		log.Println("WARNING: SMTP not configured, skipping email send")
		return nil
	}

	// If no recipient specified, skip
	if to == "" {
		log.Println("WARNING: No recipient email specified, skipping email send")
		return nil
	}

	// SMTP authentication
	auth := smtp.PlainAuth("", c.config.SMTPUsername, c.config.SMTPPassword, c.config.SMTPHost)

	// Build email message
	from := c.config.SMTPFromEmail
	fromName := c.config.SMTPFromName

	// Email headers
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", fromName, from)
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/plain; charset=\"UTF-8\""

	// Build message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Send email
	addr := fmt.Sprintf("%s:%s", c.config.SMTPHost, c.config.SMTPPort)
	err := smtp.SendMail(addr, auth, from, []string{to}, []byte(message))
	if err != nil {
		return fmt.Errorf("smtp error: %w", err)
	}

	log.Printf("Raw email sent to %s", to)
	return nil
}

// SendHTMLEmail sends an HTML email (for formatted alerts)
func (c *Client) SendHTMLEmail(to, subject, htmlBody string) error {
	// Check if SMTP is configured
	if c.config.SMTPUsername == "" || c.config.SMTPPassword == "" {
		log.Println("WARNING: SMTP not configured, skipping email send")
		return nil
	}

	// If no recipient specified, skip
	if to == "" {
		log.Println("WARNING: No recipient email specified, skipping email send")
		return nil
	}

	// SMTP authentication
	auth := smtp.PlainAuth("", c.config.SMTPUsername, c.config.SMTPPassword, c.config.SMTPHost)

	// Build email message
	from := c.config.SMTPFromEmail
	fromName := c.config.SMTPFromName

	// Email headers
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", fromName, from)
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""

	// Build message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + htmlBody

	// Send email
	addr := fmt.Sprintf("%s:%s", c.config.SMTPHost, c.config.SMTPPort)
	err := smtp.SendMail(addr, auth, from, []string{to}, []byte(message))
	if err != nil {
		return fmt.Errorf("smtp error: %w", err)
	}

	log.Printf("HTML email sent to %s", to)
	return nil
}

// Helper function to format price
func formatPrice(price float64) string {
	return fmt.Sprintf("%.2f", price)
}

// Helper function to convert string to int for port
func stringToInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

// EmailLayout wraps inner content HTML in the branded Nessie Audio email shell.
// title: the heading text shown in the header (e.g., "Order Confirmed!")
// icon: an emoji/symbol shown above the title (e.g., "&#10003;", "&#128230;")
// contentHTML: the pre-rendered inner HTML to place in the content area
// isAdmin: if true, shows admin footer instead of customer footer
func EmailLayout(title, icon, contentHTML string, isAdmin bool) string {
	footerHTML := `<p style="font-weight:700;font-size:14px;color:#c0c0c0;letter-spacing:0.05em;">NESSIE AUDIO</p>
                <p>Thank you for supporting our small business!</p>
                <p style="margin-top:12px;">Questions? Contact us at <a href="mailto:nessieaudio@gmail.com" style="color:#c0c0c0;text-decoration:underline;">nessieaudio@gmail.com</a></p>`
	if isAdmin {
		footerHTML = `<p style="font-weight:700;font-size:14px;color:#c0c0c0;letter-spacing:0.05em;">NESSIE AUDIO</p>
                <p>Admin Alert &mdash; This is an automated notification from your store.</p>`
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>
        body { margin:0; padding:0; background-color:#020202; font-family:'Montserrat',Arial,Helvetica,sans-serif; color:#ffffff; -webkit-font-smoothing:antialiased; }
        .wrapper { width:100%%; background-color:#020202; padding:20px 0; }
        .container { max-width:600px; margin:0 auto; background-color:#0a0a0a; border:1px solid #2a2458; border-radius:8px; overflow:hidden; }
        .header { background-color:#1c1838; background:linear-gradient(180deg,#1c1838,#19163a); padding:30px 20px; text-align:center; border-bottom:1px solid #3d3d3d; }
        .header img { width:80px; height:80px; margin-bottom:12px; border-radius:8px; }
        .header-icon { font-size:36px; margin:8px 0; }
        .header h1 { margin:0; font-size:24px; color:#ffffff; text-transform:uppercase; letter-spacing:2px; font-weight:700; }
        .content { padding:32px 24px; color:#ffffff; font-size:15px; line-height:1.6; }
        .info-box { background-color:#141414; border:1px solid #333; border-radius:6px; padding:20px; margin:20px 0; }
        .info-box h2 { margin:0 0 15px 0; font-size:14px; color:#ffffff; text-transform:uppercase; letter-spacing:1px; font-weight:600; }
        .detail-row { padding:8px 0; border-bottom:1px solid #2a2a2a; color:#c0c0c0; }
        .detail-row:last-child { border-bottom:none; }
        .detail-row strong { color:#fff; }
        .note { background-color:#141414; border-left:3px solid #00ff88; padding:15px; margin:20px 0; font-size:14px; color:#c0c0c0; }
        .note-error { border-left-color:#ff6b6b; }
        .cta-button { display:inline-block; background-color:#00ff88; color:#0a0a0a; padding:14px 32px; text-decoration:none; border-radius:4px; font-weight:600; text-transform:uppercase; letter-spacing:0.05em; font-size:14px; }
        .items-table { width:100%%; border-collapse:collapse; margin:20px 0; }
        .items-table th { background-color:#1a1a2e; padding:12px; text-align:left; font-size:12px; text-transform:uppercase; letter-spacing:1px; color:#ffffff; border-bottom:2px solid #2a2458; font-weight:600; }
        .items-table td { padding:15px 12px; border-bottom:1px solid #2a2a2a; color:#c0c0c0; }
        .items-table tr:last-child td { border-bottom:none; }
        .item-name { font-weight:600; color:#fff; }
        .item-variant { font-size:13px; color:#888; margin-top:4px; }
        .total-row { background-color:#141414; }
        .total-row td { padding:20px 12px; color:#ffffff; font-weight:bold; font-size:18px; }
        .stock-critical { background-color:#ff6b6b; color:#fff; padding:4px 12px; border-radius:4px; font-weight:bold; font-size:13px; display:inline-block; }
        .stock-low { background-color:#ffa500; color:#fff; padding:4px 12px; border-radius:4px; font-weight:bold; font-size:13px; display:inline-block; }
        .tracking-box { background-color:#141414; border:1px solid #333; border-radius:6px; padding:20px; margin:20px 0; text-align:center; }
        .tracking-label { margin:0; color:#888; font-size:12px; text-transform:uppercase; letter-spacing:1px; font-weight:600; }
        .tracking-number { font-size:24px; font-weight:bold; color:#00ff88; margin:15px 0; letter-spacing:2px; }
        .carrier-name { font-size:16px; color:#fff; font-weight:600; margin:10px 0 5px 0; }
        .footer { background-color:#0f0a1e; background:linear-gradient(180deg,#19163a,#1c1838); padding:24px 20px; text-align:center; border-top:1px solid #3d3d3d; }
        .footer p { margin:5px 0; font-size:13px; color:#888; }
    </style>
</head>
<body>
    <div class="wrapper">
        <div class="container">
            <div class="header">
                <img src="https://www.nessieaudio.com/android-chrome-192x192.png" alt="Nessie Audio" width="80" height="80">
                <div class="header-icon">%s</div>
                <h1>%s</h1>
            </div>
            <div class="content">
                %s
            </div>
            <div class="footer">
                %s
            </div>
        </div>
    </div>
</body>
</html>`, title, icon, title, contentHTML, footerHTML)
}

// InfoBox returns a styled info box with a title and inner HTML rows.
func InfoBox(title, innerHTML string) string {
	return fmt.Sprintf(`<div class="info-box"><h2>%s</h2>%s</div>`, title, innerHTML)
}

// DetailRow returns a label/value row for use inside InfoBox.
func DetailRow(label, value string) string {
	return fmt.Sprintf(`<div class="detail-row"><span>%s</span> <strong>%s</strong></div>`, label, value)
}

// NoteBox returns a styled note callout. Set isError=true for red accent.
func NoteBox(innerHTML string, isError bool) string {
	class := "note"
	if isError {
		class = "note note-error"
	}
	return fmt.Sprintf(`<div class="%s">%s</div>`, class, innerHTML)
}

// CTAButton returns a styled call-to-action button.
func CTAButton(text, href string) string {
	return fmt.Sprintf(`<div style="text-align:center;margin:24px 0;"><a href="%s" class="cta-button">%s</a></div>`, href, text)
}
