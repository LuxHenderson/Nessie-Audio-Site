package inventory

import (
	"bytes"
	"fmt"
	"html/template"
	"log"

	"github.com/nessieaudio/ecommerce-backend/internal/config"
	"github.com/nessieaudio/ecommerce-backend/internal/services/email"
)

// AlertService handles low-stock alerts
type AlertService struct {
	inventoryService *Service
	emailClient      *email.Client
	config           *config.Config
}

// NewAlertService creates a new alert service
func NewAlertService(inventoryService *Service, emailClient *email.Client, cfg *config.Config) *AlertService {
	return &AlertService{
		inventoryService: inventoryService,
		emailClient:      emailClient,
		config:           cfg,
	}
}

// CheckAndSendLowStockAlerts checks for low stock items and sends alert email
func (a *AlertService) CheckAndSendLowStockAlerts() error {
	// Get all low stock items
	lowStockItems, err := a.inventoryService.GetLowStockItems()
	if err != nil {
		return fmt.Errorf("get low stock items: %w", err)
	}

	// If no low stock items, nothing to do
	if len(lowStockItems) == 0 {
		log.Println("No low stock items found")
		return nil
	}

	// Send alert email
	if err := a.sendLowStockAlert(lowStockItems); err != nil {
		return fmt.Errorf("send low stock alert: %w", err)
	}

	log.Printf("Low stock alert sent for %d items", len(lowStockItems))
	return nil
}

// sendLowStockAlert sends an email alert for low stock items
func (a *AlertService) sendLowStockAlert(items []LowStockItem) error {
	// Generate HTML email body
	htmlBody, err := a.generateLowStockAlertHTML(items)
	if err != nil {
		return fmt.Errorf("generate HTML: %w", err)
	}

	subject := fmt.Sprintf("⚠️ Low Stock Alert - %d Items Need Restocking", len(items))

	// Send to admin email
	adminEmail := a.config.AdminEmail
	if adminEmail == "" {
		log.Println("WARNING: No admin email configured, skipping low stock alert")
		return nil
	}

	return a.emailClient.SendHTMLEmail(adminEmail, subject, htmlBody)
}

// generateLowStockAlertHTML generates HTML for low stock alert email
func (a *AlertService) generateLowStockAlertHTML(items []LowStockItem) (string, error) {
	tmpl := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Low Stock Alert</title>
    <style>
        body {
            font-family: 'Montserrat', Arial, sans-serif;
            background-color: #020202;
            color: #ffffff;
            margin: 0;
            padding: 0;
        }
        .container {
            max-width: 700px;
            margin: 0 auto;
            background-color: #0f0f0f;
            border: 1px solid #2a2a2a;
            border-radius: 8px;
            overflow: hidden;
        }
        .header {
            background: linear-gradient(135deg, #1a1a1a 0%, #2a2a2a 100%);
            padding: 40px 20px;
            text-align: center;
            border-bottom: 2px solid #ff6b6b;
        }
        .header h1 {
            margin: 0;
            font-family: 'Montserrat', sans-serif;
            font-size: 28px;
            color: #ff6b6b;
            text-transform: uppercase;
            letter-spacing: 2px;
            font-weight: 700;
        }
        .alert-icon {
            font-size: 48px;
            margin-bottom: 10px;
        }
        .content {
            padding: 40px 20px;
        }
        .intro {
            font-size: 16px;
            margin-bottom: 30px;
            color: #e0e0e0;
        }
        .items-table {
            width: 100%;
            border-collapse: collapse;
            margin: 20px 0;
            background-color: #1a1a1a;
            border-radius: 6px;
            overflow: hidden;
        }
        .items-table thead {
            background-color: #252525;
        }
        .items-table th {
            padding: 15px;
            text-align: left;
            font-size: 12px;
            text-transform: uppercase;
            letter-spacing: 1px;
            color: #ffffff;
            border-bottom: 2px solid rgba(255, 255, 255, 0.1);
            font-weight: 600;
        }
        .items-table td {
            padding: 15px;
            border-bottom: 1px solid #333;
            color: #e0e0e0;
        }
        .items-table tr:last-child td {
            border-bottom: none;
        }
        .items-table tr:hover {
            background-color: #252525;
        }
        .product-name {
            font-weight: 600;
            color: #fff;
        }
        .variant-name {
            font-size: 13px;
            color: #999;
            margin-top: 4px;
        }
        .stock-critical {
            background-color: #ff6b6b;
            color: #fff;
            padding: 4px 12px;
            border-radius: 4px;
            font-weight: bold;
            font-size: 13px;
            display: inline-block;
        }
        .stock-low {
            background-color: #ffa500;
            color: #fff;
            padding: 4px 12px;
            border-radius: 4px;
            font-weight: bold;
            font-size: 13px;
            display: inline-block;
        }
        .note {
            background-color: #252525;
            border-left: 3px solid #ff6b6b;
            padding: 15px;
            margin: 30px 0;
            font-size: 14px;
            color: #ccc;
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
    </style>
</head>
<body>
    <div class="container">
        <!-- Header -->
        <div class="header">
            <div class="alert-icon">⚠️</div>
            <h1>Low Stock Alert</h1>
        </div>

        <!-- Content -->
        <div class="content">
            <p class="intro">
                <strong>Action Required:</strong> The following {{len .}} item(s) are running low on stock and need restocking soon.
            </p>

            <!-- Items Table -->
            <table class="items-table">
                <thead>
                    <tr>
                        <th>Product / Variant</th>
                        <th style="text-align: center;">Current Stock</th>
                        <th style="text-align: center;">Threshold</th>
                        <th style="text-align: center;">Status</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .}}
                    <tr>
                        <td>
                            <div class="product-name">{{.ProductName}}</div>
                            <div class="variant-name">{{.VariantName}}</div>
                        </td>
                        <td style="text-align: center; font-weight: bold;">{{.StockQuantity}}</td>
                        <td style="text-align: center;">{{.LowStockThreshold}}</td>
                        <td style="text-align: center;">
                            {{if eq .StockQuantity 0}}
                                <span class="stock-critical">OUT OF STOCK</span>
                            {{else if le .StockQuantity 2}}
                                <span class="stock-critical">CRITICAL</span>
                            {{else}}
                                <span class="stock-low">LOW</span>
                            {{end}}
                        </td>
                    </tr>
                    {{end}}
                </tbody>
            </table>

            <!-- Note -->
            <div class="note">
                <strong>Recommended Actions:</strong><br>
                • Review stock levels and place restock orders<br>
                • Consider temporarily disabling low-stock variants<br>
                • Update inventory thresholds if needed
            </div>
        </div>

        <!-- Footer -->
        <div class="footer">
            <p><strong>Nessie Audio Inventory System</strong></p>
            <p>This is an automated alert. Check your inventory management dashboard for details.</p>
        </div>
    </div>
</body>
</html>
`

	// Parse template
	t, err := template.New("lowStockAlert").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	// Execute template
	var buf bytes.Buffer
	if err := t.Execute(&buf, items); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}

// SendImmediateLowStockAlert sends an alert immediately after stock is deducted (if it crosses threshold)
func (a *AlertService) SendImmediateLowStockAlert(variantID, variantName, productName string, stockQuantity, threshold int) error {
	item := LowStockItem{
		VariantID:         variantID,
		VariantName:       variantName,
		ProductName:       productName,
		StockQuantity:     stockQuantity,
		LowStockThreshold: threshold,
	}

	htmlBody, err := a.generateLowStockAlertHTML([]LowStockItem{item})
	if err != nil {
		return fmt.Errorf("generate HTML: %w", err)
	}

	subject := fmt.Sprintf("⚠️ Low Stock Alert - %s (%s)", productName, variantName)

	adminEmail := a.config.AdminEmail
	if adminEmail == "" {
		log.Println("WARNING: No admin email configured, skipping immediate low stock alert")
		return nil
	}

	return a.emailClient.SendHTMLEmail(adminEmail, subject, htmlBody)
}
