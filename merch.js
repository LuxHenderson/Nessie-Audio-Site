/**
 * merch.js
 * JavaScript module for Nessie Audio Merch page
 * 
 * This file is designed to integrate with a Golang backend for eCommerce functionality.
 * Currently contains placeholder functions that can be connected to your API endpoints.
 */

// ========== CONFIGURATION ==========
// Backend API endpoint - update this with your Golang server URL
const API_BASE_URL = 'http://localhost:8080/api'; // Example: Change to your deployed backend URL
const PRODUCTS_ENDPOINT = `${API_BASE_URL}/products`;
const CART_ENDPOINT = `${API_BASE_URL}/cart`;

// ========== PRODUCT DATA STRUCTURE ==========
/**
 * Expected product object structure from backend:
 * {
 *   id: string | number,
 *   name: string,
 *   description: string,
 *   price: number,
 *   imageUrl: string,
 *   category: string (optional),
 *   stock: number (optional),
 *   featured: boolean (optional)
 * }
 */

// ========== PLACEHOLDER PRODUCT DATA ==========
// This static data can be replaced with dynamic API calls
const placeholderProducts = [
  {
    id: '001',
    name: 'Nessie Audio Classic Tee',
    description: 'Premium cotton t-shirt with the iconic Nessie Audio logo. Comfortable fit for studio sessions or casual wear.',
    price: 29.99,
    imageUrl: 'https://via.placeholder.com/400x400/1a1a1a/ffffff?text=Nessie+Audio+Tee'
  },
  {
    id: '002',
    name: 'Studio Session Hoodie',
    description: 'Heavyweight hoodie perfect for late-night studio sessions. Features embroidered logo and kangaroo pocket.',
    price: 54.99,
    imageUrl: 'https://via.placeholder.com/400x400/1a1a1a/ffffff?text=Studio+Hoodie'
  },
  {
    id: '003',
    name: 'Nessie Snapback Hat',
    description: 'Classic snapback with embroidered Nessie logo. Adjustable fit with structured front panels.',
    price: 24.99,
    imageUrl: 'https://via.placeholder.com/400x400/1a1a1a/ffffff?text=Snapback+Hat'
  },
  {
    id: '004',
    name: 'Vinyl Sticker Pack',
    description: 'Weatherproof vinyl stickers featuring various Nessie Audio designs. Perfect for laptops, instruments, and gear.',
    price: 8.99,
    imageUrl: 'https://via.placeholder.com/400x400/1a1a1a/ffffff?text=Vinyl+Sticker'
  },
  {
    id: '005',
    name: 'Canvas Tote Bag',
    description: 'Heavy-duty canvas tote with screen-printed Nessie logo. Perfect for carrying gear, records, or daily essentials.',
    price: 19.99,
    imageUrl: 'https://via.placeholder.com/400x400/1a1a1a/ffffff?text=Tote+Bag'
  },
  {
    id: '006',
    name: 'Limited Edition Poster Pack',
    description: 'Set of three 18x24" posters featuring exclusive Nessie Audio artwork. Printed on premium matte paper.',
    price: 34.99,
    imageUrl: 'https://via.placeholder.com/400x400/1a1a1a/ffffff?text=Poster+Pack'
  }
];

// ========== FETCH PRODUCTS FROM BACKEND ==========
/**
 * Fetch products from Golang backend API
 * @returns {Promise<Array>} Array of product objects
 */
async function fetchProducts() {
  try {
    // TODO: Uncomment when backend is ready
    // const response = await fetch(PRODUCTS_ENDPOINT);
    // if (!response.ok) {
    //   throw new Error(`HTTP error! status: ${response.status}`);
    // }
    // const products = await response.json();
    // return products;

    // For now, return placeholder data
    return new Promise((resolve) => {
      setTimeout(() => resolve(placeholderProducts), 100);
    });
  } catch (error) {
    console.error('Error fetching products:', error);
    return [];
  }
}

// ========== RENDER PRODUCTS DYNAMICALLY ==========
/**
 * Generate HTML for a single product item
 * @param {Object} product - Product object
 * @returns {string} HTML string for product card
 */
function createProductHTML(product) {
  return `
    <article class="merch-item">
      <div class="merch-image-container">
        <img src="${product.imageUrl}" 
             alt="${product.name}" 
             class="merch-image">
      </div>
      <div class="merch-details">
        <h3 class="merch-title">${product.name}</h3>
        <p class="merch-description">${product.description}</p>
        <p class="merch-price">$${product.price.toFixed(2)}</p>
        <button class="merch-buy-btn" 
                data-product-id="${product.id}" 
                aria-label="Add ${product.name} to cart">
          Buy Now
        </button>
      </div>
    </article>
  `;
}

/**
 * Render all products to the merch grid
 * @param {Array} products - Array of product objects
 */
function renderProducts(products) {
  const grid = document.querySelector('.merch-grid');
  
  if (!grid) {
    console.error('Merch grid element not found');
    return;
  }

  if (products.length === 0) {
    grid.innerHTML = '<div class="merch-empty">No products available at this time.</div>';
    return;
  }

  // Generate HTML for all products
  const productsHTML = products.map(createProductHTML).join('');
  grid.innerHTML = productsHTML;

  // Attach event listeners to buy buttons after rendering
  attachBuyButtonListeners();
}

// ========== SHOPPING CART FUNCTIONALITY ==========
/**
 * Add product to shopping cart
 * @param {string} productId - ID of the product to add
 */
async function addToCart(productId) {
  try {
    console.log(`Adding product ${productId} to cart`);
    
    // TODO: Implement backend cart API call
    // const response = await fetch(CART_ENDPOINT, {
    //   method: 'POST',
    //   headers: {
    //     'Content-Type': 'application/json',
    //   },
    //   body: JSON.stringify({
    //     productId: productId,
    //     quantity: 1
    //   })
    // });
    // 
    // if (!response.ok) {
    //   throw new Error(`HTTP error! status: ${response.status}`);
    // }
    // 
    // const result = await response.json();
    // console.log('Product added to cart:', result);

    // Placeholder feedback for now
    showNotification(`Product added to cart!`);
    
  } catch (error) {
    console.error('Error adding to cart:', error);
    showNotification('Error adding product to cart', 'error');
  }
}

/**
 * Attach click event listeners to all buy buttons
 */
function attachBuyButtonListeners() {
  const buyButtons = document.querySelectorAll('.merch-buy-btn');
  
  buyButtons.forEach(button => {
    button.addEventListener('click', (e) => {
      const productId = e.target.getAttribute('data-product-id');
      if (productId) {
        addToCart(productId);
        
        // Visual feedback - button animation
        e.target.textContent = 'Added!';
        e.target.style.background = 'rgba(100, 200, 100, 0.3)';
        
        setTimeout(() => {
          e.target.textContent = 'Buy Now';
          e.target.style.background = '';
        }, 1500);
      }
    });
  });
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

// ========== INITIALIZE ON PAGE LOAD ==========
/**
 * Initialize the merch page when DOM is ready
 */
async function initMerchPage() {
  console.log('Initializing Merch page...');
  
  // Option 1: Use static HTML (current implementation in merch.html)
  // Just attach event listeners to existing buy buttons
  attachBuyButtonListeners();
  
  // Option 2: Load products dynamically (uncomment to enable)
  // const products = await fetchProducts();
  // renderProducts(products);
  
  console.log('Merch page initialized');
}

// ========== EVENT LISTENERS ==========
// Wait for DOM to be fully loaded before initializing
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', initMerchPage);
} else {
  // DOM is already loaded
  initMerchPage();
}

// ========== EXPORT FOR TESTING (if using modules) ==========
// Uncomment if using ES6 modules
// export { fetchProducts, addToCart, renderProducts };
