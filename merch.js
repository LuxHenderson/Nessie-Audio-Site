/**
 * merch.js
 * JavaScript module for Nessie Audio Merch page
 * 
 * This file is designed to integrate with a Golang backend for eCommerce functionality.
 * Currently contains placeholder functions that can be connected to your API endpoints.
 */

// ========== CONFIGURATION ==========
// Backend API endpoint - automatically detects environment (see config.js)
// NOTE: This file requires config.js to be loaded first in the HTML
const API_BASE_URL = API_CONFIG.BASE_URL;
const PRODUCTS_ENDPOINT = API_CONFIG.PRODUCTS_ENDPOINT;

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
    const response = await fetch(PRODUCTS_ENDPOINT);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const data = await response.json();
    return data.products || [];
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
  // Determine price display
  let priceDisplay;
  if (product.min_price && product.max_price && product.min_price !== product.max_price) {
    priceDisplay = `$${product.min_price.toFixed(2)} - $${product.max_price.toFixed(2)}`;
  } else {
    priceDisplay = `$${product.price.toFixed(2)}`;
  }

  // Create alt text with product name and brief description for accessibility
  const altText = product.description
    ? `${product.name} - ${product.description.substring(0, 80)}`
    : product.name;

  // Use proper <a> link for keyboard accessibility
  return `
    <article class="merch-item">
      <a href="product-detail.html?id=${product.id}"
         class="merch-item-link"
         aria-label="View details for ${product.name}, ${priceDisplay}">
        <div class="merch-image-container">
          <img src="${product.image_url || product.imageUrl}"
               alt="${altText}"
               class="merch-image"
               loading="lazy">
        </div>
        <div class="merch-details">
          <h3 class="merch-title">${product.name}</h3>
          <p class="merch-price">${priceDisplay}</p>
        </div>
      </a>
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

  console.log('Rendering products to grid:', products);
  
  // Generate HTML for all products
  const productsHTML = products.map(createProductHTML).join('');
  console.log('Generated HTML length:', productsHTML.length);
  grid.innerHTML = productsHTML;

  console.log('Products rendered to DOM');
}

// ========== INITIALIZE ON PAGE LOAD ==========
/**
 * Initialize the merch page when DOM is ready
 */
async function initMerchPage() {
  console.log('Initializing Merch page...');
  
  // Load products dynamically from backend
  const products = await fetchProducts();
  console.log('Fetched products:', products);
  console.log('Number of products:', products.length);
  
  // Sort products in desired display order
  const productOrder = [
    'Nessie Audio Unisex t-shirt',
    'Nessie Audio Unisex Champion hoodie',
    'Nessie Audio Black Glossy Mug',
    'Hardcover bound Nessie Audio notebook',
    'Nessie Audio Eco Tote Bag',
    'Nessie Audio Bubble-free stickers'
  ];
  
  const sortedProducts = products.sort((a, b) => {
    const indexA = productOrder.indexOf(a.name);
    const indexB = productOrder.indexOf(b.name);
    return indexA - indexB;
  });
  
  renderProducts(sortedProducts);
  
  console.log('Merch page initialized');
}

// ========== NAVIGATION ==========
/**
 * Navigate to product detail page
 * @param {string} productId - ID of the product
 */
function navigateToProduct(productId) {
  window.location.href = `product-detail.html?id=${productId}`;
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
