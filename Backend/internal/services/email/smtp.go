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
	tmpl := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Order Confirmation</title>
    <style>
        body {
            font-family: 'Montserrat', Arial, sans-serif;
            background-color: #020202;
            color: #ffffff;
            margin: 0;
            padding: 0;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            background-color: #0f0f0f;
            border: 1px solid #2a2a2a;
            border-radius: 8px;
            overflow: hidden;
        }
        .header {
            background: linear-gradient(135deg, #0f0f0f 0%, #1a1a1a 100%);
            padding: 40px 20px;
            text-align: center;
            border-bottom: 2px solid rgba(255, 255, 255, 0.15);
        }
        .header h1 {
            margin: 0;
            font-family: 'Montserrat', sans-serif;
            font-size: 32px;
            color: #ffffff;
            text-transform: uppercase;
            letter-spacing: 2px;
            font-weight: 700;
        }
        .success-icon {
            font-size: 48px;
            margin-bottom: 10px;
        }
        .content {
            padding: 40px 20px;
        }
        .greeting {
            font-size: 18px;
            margin-bottom: 20px;
            color: #fff;
        }
        .order-info {
            background-color: #252525;
            border: 1px solid #333;
            border-radius: 6px;
            padding: 20px;
            margin: 20px 0;
        }
        .order-info h2 {
            margin: 0 0 15px 0;
            font-size: 16px;
            color: #ffffff;
            text-transform: uppercase;
            letter-spacing: 1px;
            font-weight: 600;
        }
        .order-detail {
            display: flex;
            justify-content: space-between;
            padding: 8px 0;
            border-bottom: 1px solid #333;
        }
        .order-detail:last-child {
            border-bottom: none;
        }
        .order-detail strong {
            color: #fff;
        }
        .items-table {
            width: 100%;
            border-collapse: collapse;
            margin: 20px 0;
        }
        .items-table th {
            background-color: #1a1a1a;
            padding: 12px;
            text-align: left;
            font-size: 12px;
            text-transform: uppercase;
            letter-spacing: 1px;
            color: #ffffff;
            border-bottom: 2px solid rgba(255, 255, 255, 0.1);
            font-weight: 600;
        }
        .items-table td {
            padding: 15px 12px;
            border-bottom: 1px solid #333;
        }
        .items-table tr:last-child td {
            border-bottom: none;
        }
        .item-name {
            font-weight: 600;
            color: #fff;
        }
        .item-variant {
            font-size: 14px;
            color: #999;
            margin-top: 4px;
        }
        .total-row {
            background-color: #252525;
            font-weight: bold;
            font-size: 18px;
        }
        .total-row td {
            padding: 20px 12px;
            color: #ffffff;
        }
        .shipping-info {
            background-color: #252525;
            border: 1px solid #333;
            border-radius: 6px;
            padding: 20px;
            margin: 20px 0;
        }
        .shipping-info h2 {
            margin: 0 0 15px 0;
            font-size: 16px;
            color: #ffffff;
            text-transform: uppercase;
            letter-spacing: 1px;
            font-weight: 600;
        }
        .shipping-address {
            color: #e0e0e0;
            line-height: 1.6;
        }
        .footer {
            background-color: #0a0a0a;
            padding: 30px 20px;
            text-align: center;
            border-top: 1px solid #333;
        }
        .footer p {
            margin: 5px 0;
            font-size: 14px;
            color: #666;
        }
        .cta-button {
            display: inline-block;
            background-color: #00ff88;
            color: #0a0a0a;
            padding: 14px 32px;
            text-decoration: none;
            border-radius: 4px;
            font-weight: bold;
            text-transform: uppercase;
            letter-spacing: 1px;
            margin: 20px 0;
            font-size: 14px;
        }
        .note {
            background-color: #252525;
            border-left: 3px solid #00ff88;
            padding: 15px;
            margin: 20px 0;
            font-size: 14px;
            color: #ccc;
        }
    </style>
</head>
<body>
    <div class="container">
        <!-- Header -->
        <div class="header">
            <div class="success-icon">✓</div>
            <h1>Order Confirmed!</h1>
        </div>

        <!-- Content -->
        <div class="content">
            <p class="greeting">Hey {{.CustomerName}},</p>
            <p>Thank you for your order! Your payment has been processed successfully, and we're getting your items ready to ship.</p>

            <!-- Order Information -->
            <div class="order-info">
                <h2>Order Details</h2>
                <div class="order-detail">
                    <span>Order Number:</span>
                    <strong>#{{.OrderID}}</strong>
                </div>
                <div class="order-detail">
                    <span>Email:</span>
                    <strong>{{.CustomerEmail}}</strong>
                </div>
            </div>

            <!-- Order Items -->
            <table class="items-table">
                <thead>
                    <tr>
                        <th>Item</th>
                        <th style="text-align: center;">Qty</th>
                        <th style="text-align: right;">Price</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .Items}}
                    <tr>
                        <td>
                            <div class="item-name">{{.ProductName}}</div>
                            {{if .VariantName}}
                            <div class="item-variant">{{.VariantName}}</div>
                            {{end}}
                        </td>
                        <td style="text-align: center;">{{.Quantity}}</td>
                        <td style="text-align: right;">${{printf "%.2f" .TotalPrice}}</td>
                    </tr>
                    {{end}}
                    <tr class="total-row">
                        <td colspan="2">Total</td>
                        <td style="text-align: right;">${{printf "%.2f" .Total}}</td>
                    </tr>
                </tbody>
            </table>

            <!-- Shipping Information -->
            <div class="shipping-info">
                <h2>Shipping To</h2>
                <div class="shipping-address">
                    <strong>{{.ShippingInfo.Name}}</strong><br>
                    {{.ShippingInfo.Address}}<br>
                    {{.ShippingInfo.City}}, {{.ShippingInfo.State}} {{.ShippingInfo.Zip}}<br>
                    {{.ShippingInfo.Country}}
                </div>
            </div>

            <!-- Note -->
            <div class="note">
                <strong>What's Next?</strong><br>
                Your order will be fulfilled by our print-on-demand partner. You'll receive a shipping confirmation email with tracking information once your items are on their way (typically within 2-5 business days).
            </div>

            <!-- Call to Action -->
            <div style="text-align: center;">
                <a href="https://nessieaudio.com/merch.html" class="cta-button">Continue Shopping</a>
            </div>
        </div>

        <!-- Footer -->
        <div class="footer">
            <p><strong>Nessie Audio</strong></p>
            <p>Thank you for supporting independent music!</p>
            <p style="margin-top: 20px;">Questions? Reply to this email or contact us at nessieaudio@gmail.com</p>
        </div>
    </div>
</body>
</html>
`

	// Parse template
	t, err := template.New("orderConfirmation").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// SendShippingNotification sends a shipping/tracking update email
func (c *Client) SendShippingNotification(customerEmail, orderID, trackingNumber, trackingURL string) error {
	subject := fmt.Sprintf("Your Nessie Audio Order Has Shipped! #%s", orderID)

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Shipping Update</title>
    <style>
        body { font-family: 'Inter', Arial, sans-serif; background-color: #0a0a0a; color: #e0e0e0; margin: 0; padding: 0; }
        .container { max-width: 600px; margin: 0 auto; background-color: #1a1a1a; border: 1px solid #333; border-radius: 8px; overflow: hidden; }
        .header { background: linear-gradient(135deg, #1a1a1a 0%%, #2a2a2a 100%%); padding: 40px 20px; text-align: center; border-bottom: 2px solid #00ff88; }
        .header h1 { margin: 0; font-family: 'Oswald', sans-serif; font-size: 32px; color: #00ff88; text-transform: uppercase; letter-spacing: 2px; }
        .ship-icon { font-size: 48px; margin-bottom: 10px; }
        .content { padding: 40px 20px; }
        .tracking-box { background-color: #252525; border: 1px solid #333; border-radius: 6px; padding: 20px; margin: 20px 0; text-align: center; }
        .tracking-number { font-size: 24px; font-weight: bold; color: #00ff88; margin: 15px 0; letter-spacing: 2px; }
        .cta-button { display: inline-block; background-color: #00ff88; color: #0a0a0a; padding: 14px 32px; text-decoration: none; border-radius: 4px; font-weight: bold; text-transform: uppercase; letter-spacing: 1px; margin: 20px 0; font-size: 14px; }
        .footer { background-color: #0a0a0a; padding: 30px 20px; text-align: center; border-top: 1px solid #333; }
        .footer p { margin: 5px 0; font-size: 14px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="ship-icon">📦</div>
            <h1>Your Order Has Shipped!</h1>
        </div>
        <div class="content">
            <p style="font-size: 18px; color: #fff;">Great news! Your Nessie Audio order is on its way.</p>
            <div class="tracking-box">
                <p style="margin: 0; color: #999; font-size: 14px; text-transform: uppercase; letter-spacing: 1px;">Tracking Number</p>
                <div class="tracking-number">%s</div>
                <a href="%s" class="cta-button">Track Your Package</a>
            </div>
            <p style="color: #ccc;">Your order <strong>#%s</strong> has been shipped and is on its way to you. Click the button above to track your package in real-time.</p>
        </div>
        <div class="footer">
            <p><strong>Nessie Audio</strong></p>
            <p>Questions? Reply to this email or contact us at nessieaudio@gmail.com</p>
        </div>
    </div>
</body>
</html>
`, trackingNumber, trackingURL, orderID)

	to := []string{customerEmail}
	if err := c.sendEmail(to, subject, htmlBody); err != nil {
		return fmt.Errorf("failed to send shipping notification: %w", err)
	}

	log.Printf("Shipping notification sent to %s for order %s", customerEmail, orderID)
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
