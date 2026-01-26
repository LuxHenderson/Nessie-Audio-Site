// product-detail.js
// Requires config.js to be loaded first

const API_BASE_URL = API_CONFIG.BASE_URL;
const PRODUCTS_ENDPOINT = API_CONFIG.PRODUCTS_ENDPOINT;

function getProductIdFromURL() {
  const urlParams = new URLSearchParams(window.location.search);
  return urlParams.get('id');
}

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

// Fallback when individual product endpoint doesn't exist
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

function formatDescription(description) {
  if (!description) return '';

  const paragraphs = description.split('\n\n').filter(p => p.trim());

  return paragraphs.map(p => {
    const trimmed = p.trim();

    if (trimmed.includes('\n-')) {
      const lines = trimmed.split('\n');
      const listItems = [];
      let regularText = [];

      lines.forEach(line => {
        if (line.trim().startsWith('-')) {
          if (regularText.length > 0) {
            listItems.push(`<p>${regularText.join(' ')}</p>`);
            regularText = [];
          }
          listItems.push(`<li>${line.trim().substring(1).trim()}</li>`);
        } else if (line.trim()) {
          regularText.push(line.trim());
        }
      });

      if (regularText.length > 0) {
        listItems.push(`<p>${regularText.join(' ')}</p>`);
      }

      const listHTML = listItems.map(item => {
        if (item.startsWith('<li>')) {
          return item;
        }
        return item;
      }).join('');

      const parts = listHTML.split('<li>');
      if (parts.length > 1) {
        return parts[0] + '<ul>' + parts.slice(1).map(p => '<li>' + p).join('') + '</ul>';
      }
      return listHTML;
    }

    return `<p>${trimmed}</p>`;
  }).join('<br>');
}

function renderProductDetail(product) {
  const container = document.querySelector('.product-detail-container');

  if (!container) {
    console.error('Product detail container not found');
    return;
  }

  const html = `
    <div class="product-detail">
      <div class="product-detail-image">
        <img src="${product.image_url || product.imageUrl}"
             alt="${product.name}"
             class="product-main-image"
             loading="eager">
      </div>

      <div class="product-detail-info">
        <h1 class="product-detail-title">${product.name}</h1>

        <div class="product-detail-price">
          <span class="price-amount" id="current-price">$${parseFloat(product.price).toFixed(2)}</span>
          <span class="price-currency">${product.currency || 'USD'}</span>
        </div>

        ${renderVariantsSection(product)}

        <div class="product-detail-description">
          <h3>About this product</h3>
          <div>${formatDescription(product.description) || '<p>Premium Nessie Audio merchandise, crafted for creators and fans.</p>'}</div>
        </div>

        <div class="product-availability">
          <span class="availability-badge in-stock">In Stock</span>
        </div>

        <div class="product-actions">
          <button class="btn-add-to-cart" id="add-to-cart-btn" data-product-id="${product.id}">
            Add to Cart
          </button>
          <button class="btn-buy-now" id="buy-now-btn" data-product-id="${product.id}">
            Buy Now
          </button>
        </div>

        <div class="product-meta">
          <p><strong>Category:</strong> ${product.category || 'Merchandise'}</p>
          ${product.printful_id ? `<p><strong>SKU:</strong> ${product.printful_id}</p>` : ''}
        </div>
      </div>
    </div>
  `;

  container.innerHTML = html;
  attachProductDetailListeners(product);
}

function sortVariantsBySize(variants) {
  const clothingSizeOrder = ['XS', 'S', 'M', 'L', 'XL', '2XL', '3XL', '4XL', '5XL'];

  return variants.sort((a, b) => {
    const sizeA = a.size.trim();
    const sizeB = b.size.trim();

    const aIsClothing = clothingSizeOrder.includes(sizeA);
    const bIsClothing = clothingSizeOrder.includes(sizeB);

    if (aIsClothing && bIsClothing) {
      return clothingSizeOrder.indexOf(sizeA) - clothingSizeOrder.indexOf(sizeB);
    }

    // Mugs use oz
    if (sizeA.includes('oz') && sizeB.includes('oz')) {
      const ozA = parseInt(sizeA.match(/(\d+)\s*oz/)[1]);
      const ozB = parseInt(sizeB.match(/(\d+)\s*oz/)[1]);
      return ozA - ozB;
    }

    // Stickers use dimensions
    if (sizeA.includes('″') && sizeB.includes('″')) {
      const dimA = parseFloat(sizeA.match(/(\d+\.?\d*)″/)[1]);
      const dimB = parseFloat(sizeB.match(/(\d+\.?\d*)″/)[1]);
      return dimA - dimB;
    }

    return sizeA.localeCompare(sizeB);
  });
}

function renderVariantsSection(product) {
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

  // Notebooks and totes use color variants; other products use size
  const hasColorVariants = product.name.toLowerCase().includes('notebook') ||
                           product.name.toLowerCase().includes('tote') ||
                           product.name.toLowerCase().includes('bag');
  const variantLabel = hasColorVariants ? 'Color:' : 'Size:';

  const sizeOptions = product.variants.map(v => {
    // Variant names follow pattern "Product Name / Size"
    const size = v.name.split(' / ')[1] || v.size || v.name;
    return {
      size: size,
      price: v.price,
      id: v.id,
      available: v.available
    };
  });

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

function attachProductDetailListeners(product) {
  const addToCartBtn = document.getElementById('add-to-cart-btn');
  if (addToCartBtn) {
    addToCartBtn.addEventListener('click', () => {
      handleAddToCart(product);
    });
  }

  const buyNowBtn = document.getElementById('buy-now-btn');
  if (buyNowBtn) {
    buyNowBtn.addEventListener('click', () => {
      handleBuyNow(product);
    });
  }

  const variantSelect = document.getElementById('variant-select');
  if (variantSelect) {
    variantSelect.addEventListener('change', (e) => {
      const selectedOption = e.target.options[e.target.selectedIndex];
      const newPrice = selectedOption.getAttribute('data-price');

      if (newPrice) {
        const priceElement = document.getElementById('current-price');
        if (priceElement) {
          priceElement.textContent = `$${parseFloat(newPrice).toFixed(2)}`;
        }
        console.log('Size selected:', selectedOption.text, 'Price:', newPrice);
      }
    });
  }
}

function handleAddToCart(product) {
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

  const quantityInput = document.getElementById('quantity-input');
  const quantity = quantityInput ? parseInt(quantityInput.value) : 1;

  // API returns either image_url or imageUrl depending on source
  const productImage = product.image_url || product.imageUrl || (product.images && product.images[0]) || '';

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

function handleBuyNow(product) {
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

  const quantityInput = document.getElementById('quantity-input');
  const quantity = quantityInput ? parseInt(quantityInput.value) : 1;

  // API returns either image_url or imageUrl depending on source
  const productImage = product.image_url || product.imageUrl || (product.images && product.images[0]) || '';

  // Skip notification since redirect provides immediate feedback
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

    window.location.href = 'cart.html';
  } else {
    console.error('Cart not initialized');
    showNotification('Cart not available', 'error');
  }
}

function showNotification(message, type = 'success') {
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

  // Inject animation keyframes once
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

  // Auto-dismiss after 3s
  setTimeout(() => {
    notification.style.animation = 'slideIn 0.3s ease reverse';
    setTimeout(() => notification.remove(), 300);
  }, 3000);
}

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

async function initProductDetailPage() {
  console.log('Initializing Product Detail page...');

  const productId = getProductIdFromURL();

  if (!productId) {
    console.error('No product ID provided');
    showProductNotFound();
    return;
  }

  console.log('Loading product:', productId);

  // Try direct fetch first, fall back to list search
  let product = await fetchProduct(productId);

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

  renderProductDetail(product);
  document.title = `${product.name} - Nessie Audio`;
  updateMetaTags(product);

  console.log('Product Detail page initialized');
}

function updateMetaTags(product) {
  const productUrl = `https://nessieaudio.com/product-detail.html?id=${product.id}`;
  const productImage = product.image_url || product.imageUrl || 'https://nessieaudio.com/Nessie Audio 2026.jpg';
  const productDescription = product.description
    ? product.description.substring(0, 150).replace(/\n/g, ' ').trim() + '...'
    : `Shop ${product.name} at Nessie Audio - Premium merchandise with unique designs.`;

  updateMetaTag('name', 'description', productDescription);
  updateMetaTag('property', 'og:url', productUrl);
  updateMetaTag('property', 'og:title', `${product.name} - Nessie Audio Merch`);
  updateMetaTag('property', 'og:description', productDescription);
  updateMetaTag('property', 'og:image', productImage);
  updateMetaTag('property', 'twitter:url', productUrl);
  updateMetaTag('property', 'twitter:title', `${product.name} - Nessie Audio Merch`);
  updateMetaTag('property', 'twitter:description', productDescription);
  updateMetaTag('property', 'twitter:image', productImage);

  console.log('Meta tags updated for product:', product.name);
}

function updateMetaTag(attribute, value, content) {
  let tag = document.querySelector(`meta[${attribute}="${value}"]`);

  if (tag) {
    tag.setAttribute('content', content);
  } else {
    tag = document.createElement('meta');
    tag.setAttribute(attribute, value);
    tag.setAttribute('content', content);
    document.head.appendChild(tag);
  }
}

// Wait for DOM before initializing
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', initProductDetailPage);
} else {
  initProductDetailPage();
}
