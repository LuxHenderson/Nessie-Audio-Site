/**
 * product-detail.js
 * JavaScript module for Nessie Audio product detail page
 * 
 * Handles loading and displaying individual product details including
 * variants, pricing, images, and add-to-cart functionality.
 */

// ========== CONFIGURATION ==========
const API_BASE_URL = 'http://localhost:8080/api/v1';
const PRODUCTS_ENDPOINT = `${API_BASE_URL}/products`;

// ========== GET PRODUCT ID FROM URL ==========
/**
 * Extract product ID from URL query parameters
 * @returns {string|null} Product ID or null if not found
 */
function getProductIdFromURL() {
  const urlParams = new URLSearchParams(window.location.search);
  return urlParams.get('id');
}

// ========== FETCH PRODUCT DATA ==========
/**
 * Fetch individual product details from backend
 * @param {string} productId - The product ID
 * @returns {Promise<Object|null>} Product object or null
 */
async function fetchProduct(productId) {
  try {
    const response = await fetch(`${PRODUCTS_ENDPOINT}/${productId}`);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const data = await response.json();
    return data;
  } catch (error) {
    console.error('Error fetching product:', error);
    return null;
  }
}

/**
 * Fetch all products and find the one matching the ID
 * (Fallback if individual product endpoint doesn't exist)
 * @param {string} productId - The product ID
 * @returns {Promise<Object|null>} Product object or null
 */
async function fetchProductFromList(productId) {
  try {
    const response = await fetch(PRODUCTS_ENDPOINT);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const data = await response.json();
    const products = data.products || [];
    return products.find(p => p.id === productId) || null;
  } catch (error) {
    console.error('Error fetching products:', error);
    return null;
  }
}

// ========== FORMAT DESCRIPTION ==========
/**
 * Format product description by converting line breaks to paragraph tags
 * @param {string} description - Raw description text
 * @returns {string} HTML formatted description
 */
function formatDescription(description) {
  if (!description) return '';
  
  // Split by double line breaks to create paragraphs
  const paragraphs = description.split('\n\n').filter(p => p.trim());
  
  return paragraphs.map(p => {
    const trimmed = p.trim();
    
    // Check if paragraph contains bullet points (lines starting with -)
    if (trimmed.includes('\n-')) {
      const lines = trimmed.split('\n');
      const listItems = [];
      let regularText = [];
      
      lines.forEach(line => {
        if (line.trim().startsWith('-')) {
          // If we have regular text before bullets, add it as a paragraph
          if (regularText.length > 0) {
            listItems.push(`<p>${regularText.join(' ')}</p>`);
            regularText = [];
          }
          listItems.push(`<li>${line.trim().substring(1).trim()}</li>`);
        } else if (line.trim()) {
          regularText.push(line.trim());
        }
      });
      
      // Add any remaining regular text
      if (regularText.length > 0) {
        listItems.push(`<p>${regularText.join(' ')}</p>`);
      }
      
      // Wrap list items in <ul> tags
      const listHTML = listItems.map(item => {
        if (item.startsWith('<li>')) {
          return item;
        }
        return item;
      }).join('');
      
      // Find where the list items start and wrap them
      const parts = listHTML.split('<li>');
      if (parts.length > 1) {
        return parts[0] + '<ul>' + parts.slice(1).map(p => '<li>' + p).join('') + '</ul>';
      }
      return listHTML;
    }
    
    return `<p>${trimmed}</p>`;
  }).join('<br>');
}

// ========== RENDER PRODUCT DETAIL ==========
/**
 * Render the complete product detail view
 * @param {Object} product - Product object with all details
 */
function renderProductDetail(product) {
  const container = document.querySelector('.product-detail-container');
  
  if (!container) {
    console.error('Product detail container not found');
    return;
  }

  // Build the product detail HTML
  const html = `
    <div class="product-detail">
      <!-- Product Image Section -->
      <div class="product-detail-image">
        <img src="${product.image_url || product.imageUrl}" 
             alt="${product.name}" 
             class="product-main-image">
      </div>

      <!-- Product Info Section -->
      <div class="product-detail-info">
        <h1 class="product-detail-title">${product.name}</h1>
        
        <div class="product-detail-price">
          <span class="price-amount" id="current-price">$${parseFloat(product.price).toFixed(2)}</span>
          <span class="price-currency">${product.currency || 'USD'}</span>
        </div>

        <!-- Variants Selection (if available) -->
        ${renderVariantsSection(product)}

        <!-- Product Description -->
        <div class="product-detail-description">
          <h3>About this product</h3>
          <div>${formatDescription(product.description) || '<p>Premium Nessie Audio merchandise, crafted for creators and fans.</p>'}</div>
        </div>

        <!-- Availability Status -->
        <div class="product-availability">
          <span class="availability-badge in-stock">In Stock</span>
        </div>

        <!-- Add to Cart Section -->
        <div class="product-actions">
          <button class="btn-add-to-cart" id="add-to-cart-btn" data-product-id="${product.id}">
            Add to Cart
          </button>
          <button class="btn-buy-now" id="buy-now-btn" data-product-id="${product.id}">
            Buy Now
          </button>
        </div>

        <!-- Additional Info -->
        <div class="product-meta">
          <p><strong>Category:</strong> ${product.category || 'Merchandise'}</p>
          ${product.printful_id ? `<p><strong>SKU:</strong> ${product.printful_id}</p>` : ''}
        </div>
      </div>
    </div>
  `;

  container.innerHTML = html;

  // Attach event listeners
  attachProductDetailListeners(product);
}

/**
 * Sort variants by size in logical order
 * @param {Array} variants - Array of variant objects with size property
 * @returns {Array} Sorted variants array
 */
function sortVariantsBySize(variants) {
  // Define size order for clothing
  const clothingSizeOrder = ['XS', 'S', 'M', 'L', 'XL', '2XL', '3XL', '4XL', '5XL'];
  
  return variants.sort((a, b) => {
    const sizeA = a.size.trim();
    const sizeB = b.size.trim();
    
    // Check if both are clothing sizes
    const aIsClothing = clothingSizeOrder.includes(sizeA);
    const bIsClothing = clothingSizeOrder.includes(sizeB);
    
    if (aIsClothing && bIsClothing) {
      return clothingSizeOrder.indexOf(sizeA) - clothingSizeOrder.indexOf(sizeB);
    }
    
    // Check if both contain "oz" (mugs)
    if (sizeA.includes('oz') && sizeB.includes('oz')) {
      const ozA = parseInt(sizeA.match(/(\d+)\s*oz/)[1]);
      const ozB = parseInt(sizeB.match(/(\d+)\s*oz/)[1]);
      return ozA - ozB;
    }
    
    // Check if both contain dimensions (stickers - e.g., "3″×3″")
    if (sizeA.includes('″') && sizeB.includes('″')) {
      // Extract first dimension for comparison
      const dimA = parseFloat(sizeA.match(/(\d+\.?\d*)″/)[1]);
      const dimB = parseFloat(sizeB.match(/(\d+\.?\d*)″/)[1]);
      return dimA - dimB;
    }
    
    // Default: alphabetical order
    return sizeA.localeCompare(sizeB);
  });
}

/**
 * Render variants selection section if product has variants
 * @param {Object} product - Product object
 * @returns {string} HTML string for variants section
 */
function renderVariantsSection(product) {
  // Check if product has variants
  if (!product.variants || product.variants.length === 0) {
    return `
      <div class="product-variants">
        <label for="variant-select">Size:</label>
        <select id="variant-select" class="variant-selector">
          <option value="default">One Size</option>
        </select>
      </div>
    `;
  }

  // Determine if this is a notebook or tote bag (color variants) or other product (size variants)
  const hasColorVariants = product.name.toLowerCase().includes('notebook') || 
                           product.name.toLowerCase().includes('tote') ||
                           product.name.toLowerCase().includes('bag');
  const variantLabel = hasColorVariants ? 'Color:' : 'Size:';

  // Extract unique sizes from variants
  const sizeOptions = product.variants.map(v => {
    // Extract size from variant name (e.g., "Nessie Audio Unisex t-shirt / XS" -> "XS")
    const size = v.name.split(' / ')[1] || v.size || v.name;
    return {
      size: size,
      price: v.price,
      id: v.id,
      available: v.available
    };
  });

  // Sort variants by size
  const sortedOptions = sortVariantsBySize(sizeOptions);

  const optionsHTML = sortedOptions.map(opt => 
    `<option value="${opt.id}" data-price="${opt.price}">${opt.size} - $${parseFloat(opt.price).toFixed(2)}</option>`
  ).join('');

  return `
    <div class="product-variants">
      <label for="variant-select">${variantLabel}</label>
      <select id="variant-select" class="variant-selector" data-base-price="${product.price}">
        ${optionsHTML}
      </select>
    </div>
  `;
}

/**
 * Attach event listeners for product detail page interactions
 * @param {Object} product - Product object
 */
function attachProductDetailListeners(product) {
  // Add to Cart button
  const addToCartBtn = document.getElementById('add-to-cart-btn');
  if (addToCartBtn) {
    addToCartBtn.addEventListener('click', () => {
      handleAddToCart(product);
    });
  }

  // Buy Now button
  const buyNowBtn = document.getElementById('buy-now-btn');
  if (buyNowBtn) {
    buyNowBtn.addEventListener('click', () => {
      handleBuyNow(product);
    });
  }

  // Variant selector - update price when size changes
  const variantSelect = document.getElementById('variant-select');
  if (variantSelect) {
    variantSelect.addEventListener('change', (e) => {
      const selectedOption = e.target.options[e.target.selectedIndex];
      const newPrice = selectedOption.getAttribute('data-price');
      
      if (newPrice) {
        // Update the displayed price
        const priceElement = document.getElementById('current-price');
        if (priceElement) {
          priceElement.textContent = `$${parseFloat(newPrice).toFixed(2)}`;
        }
        console.log('Size selected:', selectedOption.text, 'Price:', newPrice);
      }
    });
  }
}

// ========== CART ACTIONS (STUBBED) ==========
/**
 * Handle add to cart action
 * @param {Object} product - Product object
 */
function handleAddToCart(product) {
  // Get selected variant
  const variantSelect = document.getElementById('variant-select');
  if (!variantSelect) {
    showNotification('Please select a size', 'error');
    return;
  }

  const selectedVariantId = variantSelect.value;
  const selectedVariant = product.variants.find(v => v.id === selectedVariantId);
  
  if (!selectedVariant) {
    console.error('Selected variant not found:', {
      selectedVariantId,
      availableVariants: product.variants.map(v => ({ id: v.id, name: v.name }))
    });
    showNotification('Please select a valid size', 'error');
    return;
  }

  // Get quantity
  const quantityInput = document.getElementById('quantity-input');
  const quantity = quantityInput ? parseInt(quantityInput.value) : 1;

  // Use image_url or imageUrl from product
  const productImage = product.image_url || product.imageUrl || (product.images && product.images[0]) || '';

  // Add to cart using the global cart object
  if (window.cart) {
    cart.addItem({
      productId: product.id,
      productName: product.name,
      variantId: selectedVariant.id,
      variantName: selectedVariant.name,
      price: selectedVariant.price,
      image: productImage,
      quantity: quantity
    });
    
    showNotification(`${product.name} added to cart!`, 'success');
  } else {
    console.error('Cart not initialized');
    showNotification('Cart not available', 'error');
  }
}

/**
 * Handle buy now action
 * @param {Object} product - Product object
 */
function handleBuyNow(product) {
  // Get selected variant
  const variantSelect = document.getElementById('variant-select');
  if (!variantSelect) {
    showNotification('Please select a size', 'error');
    return;
  }

  const selectedVariantId = variantSelect.value;
  const selectedVariant = product.variants.find(v => v.id === selectedVariantId);

  if (!selectedVariant) {
    showNotification('Please select a valid size', 'error');
    return;
  }

  // Get quantity
  const quantityInput = document.getElementById('quantity-input');
  const quantity = quantityInput ? parseInt(quantityInput.value) : 1;

  // Use image_url or imageUrl from product
  const productImage = product.image_url || product.imageUrl || (product.images && product.images[0]) || '';

  // Add to cart silently (no notification)
  if (window.cart) {
    cart.addItem({
      productId: product.id,
      productName: product.name,
      variantId: selectedVariant.id,
      variantName: selectedVariant.name,
      price: selectedVariant.price,
      image: productImage,
      quantity: quantity
    });

    // Immediately redirect to cart page (no notification)
    window.location.href = 'cart.html';
  } else {
    console.error('Cart not initialized');
    showNotification('Cart not available', 'error');
  }
}

// ========== NOTIFICATION SYSTEM ==========
/**
 * Display notification message to user
 * @param {string} message - Notification message
 * @param {string} type - Notification type ('success' or 'error')
 */
function showNotification(message, type = 'success') {
  // Check if notification container exists, create if not
  let notificationContainer = document.querySelector('.notification-container');
  
  if (!notificationContainer) {
    notificationContainer = document.createElement('div');
    notificationContainer.className = 'notification-container';
    notificationContainer.style.cssText = `
      position: fixed;
      top: 100px;
      right: 20px;
      z-index: 2000;
      display: flex;
      flex-direction: column;
      gap: 10px;
    `;
    document.body.appendChild(notificationContainer);
  }

  // Create notification element
  const notification = document.createElement('div');
  notification.className = `notification notification-${type}`;
  notification.textContent = message;
  notification.style.cssText = `
    background: ${type === 'success' ? 'rgba(100, 200, 100, 0.9)' : 'rgba(200, 100, 100, 0.9)'};
    color: white;
    padding: 1rem 1.5rem;
    border-radius: 8px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
    font-size: 0.95rem;
    font-weight: 600;
    animation: slideIn 0.3s ease;
  `;

  // Add CSS animation
  if (!document.querySelector('#notification-styles')) {
    const style = document.createElement('style');
    style.id = 'notification-styles';
    style.textContent = `
      @keyframes slideIn {
        from {
          transform: translateX(400px);
          opacity: 0;
        }
        to {
          transform: translateX(0);
          opacity: 1;
        }
      }
    `;
    document.head.appendChild(style);
  }

  notificationContainer.appendChild(notification);

  // Auto-remove notification after 3 seconds
  setTimeout(() => {
    notification.style.animation = 'slideIn 0.3s ease reverse';
    setTimeout(() => notification.remove(), 300);
  }, 3000);
}

/**
 * Show error state when product not found
 */
function showProductNotFound() {
  const container = document.querySelector('.product-detail-container');
  if (container) {
    container.innerHTML = `
      <div class="product-not-found">
        <h2>Product Not Found</h2>
        <p>Sorry, we couldn't find the product you're looking for.</p>
        <a href="merch.html" class="btn">Back to Merch</a>
      </div>
    `;
  }
}

// ========== INITIALIZE PAGE ==========
/**
 * Initialize the product detail page
 */
async function initProductDetailPage() {
  console.log('Initializing Product Detail page...');
  
  // Get product ID from URL
  const productId = getProductIdFromURL();
  
  if (!productId) {
    console.error('No product ID provided');
    showProductNotFound();
    return;
  }

  console.log('Loading product:', productId);
  
  // Try to fetch individual product first
  let product = await fetchProduct(productId);
  
  // If that fails, fetch from list (fallback)
  if (!product) {
    console.log('Trying fallback method...');
    product = await fetchProductFromList(productId);
  }
  
  if (!product) {
    console.error('Product not found:', productId);
    showProductNotFound();
    return;
  }

  console.log('Product loaded:', product);
  
  // Render the product detail view
  renderProductDetail(product);
  
  // Update page title
  document.title = `${product.name} - Nessie Audio`;
  
  console.log('Product Detail page initialized');
}

// ========== EVENT LISTENERS ==========
// Wait for DOM to be fully loaded before initializing
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', initProductDetailPage);
} else {
  // DOM is already loaded
  initProductDetailPage();
}
