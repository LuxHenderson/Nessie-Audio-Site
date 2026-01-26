// merch.js
// Requires config.js to be loaded first

const API_BASE_URL = API_CONFIG.BASE_URL;
const PRODUCTS_ENDPOINT = API_CONFIG.PRODUCTS_ENDPOINT;

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

function createProductHTML(product) {
  // Show price range if variants have different prices
  let priceDisplay;
  if (product.min_price && product.max_price && product.min_price !== product.max_price) {
    priceDisplay = `$${product.min_price.toFixed(2)} - $${product.max_price.toFixed(2)}`;
  } else {
    priceDisplay = `$${product.price.toFixed(2)}`;
  }

  const altText = product.description
    ? `${product.name} - ${product.description.substring(0, 80)}`
    : product.name;

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

  grid.innerHTML = products.map(createProductHTML).join('');
}

async function initMerchPage() {
  const products = await fetchProducts();

  // Display products in curated order
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
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', initMerchPage);
} else {
  initMerchPage();
}
