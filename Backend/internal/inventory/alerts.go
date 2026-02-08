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
	innerTmpl := `
            <p style="font-size:16px;"><strong>Action Required:</strong> The following {{len .}} item(s) are running low on stock and need restocking soon.</p>

            <table class="items-table">
                <thead>
                    <tr>
                        <th>Product / Variant</th>
                        <th style="text-align:center;">Current Stock</th>
                        <th style="text-align:center;">Threshold</th>
                        <th style="text-align:center;">Status</th>
                    </tr>
                </thead>
                <tbody>
                    {{range .}}
                    <tr>
                        <td>
                            <div class="item-name">{{.ProductName}}</div>
                            <div class="item-variant">{{.VariantName}}</div>
                        </td>
                        <td style="text-align:center;font-weight:bold;">{{.StockQuantity}}</td>
                        <td style="text-align:center;">{{.LowStockThreshold}}</td>
                        <td style="text-align:center;">
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

            <div class="note note-error"><strong>Recommended Actions:</strong><br>&bull; Review stock levels and place restock orders<br>&bull; Consider temporarily disabling low-stock variants<br>&bull; Update inventory thresholds if needed</div>`

	t, err := template.New("lowStockAlert").Parse(innerTmpl)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, items); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return email.EmailLayout("Low Stock Alert", "&#9888;&#65039;", buf.String(), true), nil
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
