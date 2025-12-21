/**
 * cart-page.js
 * Handles the shopping cart page display and interactions
 */

const API_BASE_URL = 'http://localhost:8080/api/v1';
let stripeInstance = null;

// Initialize Stripe
async function initStripe() {
  try {
    const response = await fetch(`${API_BASE_URL}/config`);
    const config = await response.json();
    stripeInstance = Stripe(config.stripe_publishable_key);
  } catch (error) {
    console.error('Failed to initialize Stripe:', error);
  }
}

// ========== RENDER CART PAGE ==========
function renderCartPage() {
  const items = cart.getItems();
  const emptyCart = document.getElementById('empty-cart');
  const cartContents = document.getElementById('cart-contents');
  const cartItemsList = document.getElementById('cart-items-list');

  if (items.length === 0) {
    emptyCart.style.display = 'block';
    cartContents.style.display = 'none';
    return;
  }

  emptyCart.style.display = 'none';
  cartContents.style.display = 'grid';

  // Render each cart item
  cartItemsList.innerHTML = items.map(item => createCartItemHTML(item)).join('');

  // Update summary
  updateCartSummary();

  // Attach event listeners
  attachCartItemListeners();
}

/**
 * Create HTML for a single cart item
 * @param {Object} item - Cart item
 * @returns {string} HTML string
 */
function createCartItemHTML(item) {
  const itemTotal = (item.price * item.quantity).toFixed(2);
  
  return `
    <div class="cart-item" data-product-id="${item.productId}" data-variant-id="${item.variantId}">
      <div class="cart-item-image">
        <img src="${item.imageUrl}" alt="${item.name}">
      </div>
      
      <div class="cart-item-details">
        <h3 class="cart-item-name">${item.name}</h3>
        <p class="cart-item-variant">${item.variantName}</p>
        <p class="cart-item-price">$${item.price.toFixed(2)}</p>
      </div>

      <div class="cart-item-quantity">
        <label for="qty-${item.variantId}">Quantity:</label>
        <div class="quantity-controls">
          <button class="qty-btn qty-decrease" data-product-id="${item.productId}" data-variant-id="${item.variantId}">−</button>
          <input 
            type="number" 
            id="qty-${item.variantId}"
            class="qty-input" 
            value="${item.quantity}" 
            min="1" 
            max="99"
            data-product-id="${item.productId}"
            data-variant-id="${item.variantId}">
          <button class="qty-btn qty-increase" data-product-id="${item.productId}" data-variant-id="${item.variantId}">+</button>
        </div>
      </div>

      <div class="cart-item-total">
        <p class="item-total-label">Total:</p>
        <p class="item-total-price">$${itemTotal}</p>
      </div>

      <button class="cart-item-remove" data-product-id="${item.productId}" data-variant-id="${item.variantId}" aria-label="Remove ${item.name}">
        ✕
      </button>
    </div>
  `;
}

/**
 * Update cart summary totals
 */
function updateCartSummary() {
  const subtotal = cart.getSubtotal();
  
  document.getElementById('cart-subtotal').textContent = `$${subtotal.toFixed(2)}`;
  document.getElementById('cart-total').textContent = `$${subtotal.toFixed(2)}`;
}

/**
 * Attach event listeners to cart item controls
 */
function attachCartItemListeners() {
  // Quantity decrease buttons
  document.querySelectorAll('.qty-decrease').forEach(btn => {
    btn.addEventListener('click', (e) => {
      const productId = e.target.dataset.productId;
      const variantId = e.target.dataset.variantId;
      const input = document.querySelector(`#qty-${variantId}`);
      const currentQty = parseInt(input.value);
      
      if (currentQty > 1) {
        cart.updateQuantity(productId, variantId, currentQty - 1);
        renderCartPage();
      }
    });
  });

  // Quantity increase buttons
  document.querySelectorAll('.qty-increase').forEach(btn => {
    btn.addEventListener('click', (e) => {
      const productId = e.target.dataset.productId;
      const variantId = e.target.dataset.variantId;
      const input = document.querySelector(`#qty-${variantId}`);
      const currentQty = parseInt(input.value);
      
      if (currentQty < 99) {
        cart.updateQuantity(productId, variantId, currentQty + 1);
        renderCartPage();
      }
    });
  });

  // Quantity input direct change
  document.querySelectorAll('.qty-input').forEach(input => {
    input.addEventListener('change', (e) => {
      const productId = e.target.dataset.productId;
      const variantId = e.target.dataset.variantId;
      let newQty = parseInt(e.target.value);
      
      if (isNaN(newQty) || newQty < 1) {
        newQty = 1;
      } else if (newQty > 99) {
        newQty = 99;
      }
      
      cart.updateQuantity(productId, variantId, newQty);
      renderCartPage();
    });
  });

  // Remove buttons
  document.querySelectorAll('.cart-item-remove').forEach(btn => {
    btn.addEventListener('click', (e) => {
      const productId = e.target.dataset.productId;
      const variantId = e.target.dataset.variantId;
      
      if (confirm('Remove this item from your cart?')) {
        cart.removeItem(productId, variantId);
        renderCartPage();
      }
    });
  });
}

/**
 * Handle checkout button click
 */
async function handleCheckout() {
  const items = cart.getItems();
  
  if (items.length === 0) {
    alert('Your cart is empty.');
    return;
  }

  // Prompt for email
  const email = prompt('Please enter your email address for order confirmation:');
  if (!email || !email.includes('@')) {
    alert('A valid email address is required for checkout.');
    return;
  }

  const checkoutBtn = document.getElementById('checkout-btn');
  checkoutBtn.disabled = true;
  checkoutBtn.textContent = 'Processing...';

  try {
    // Create Stripe checkout session
    const response = await fetch(`${API_BASE_URL}/cart/checkout`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        email: email,
        items: items.map(item => ({
          product_id: item.productId,
          variant_id: item.variantId,
          quantity: item.quantity
        }))
      })
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || `Checkout failed: ${response.status}`);
    }

    const data = await response.json();
    
    if (data.session_id) {
      // Redirect to Stripe Checkout using Stripe.js
      if (!stripeInstance) {
        throw new Error('Stripe not initialized');
      }
      await stripeInstance.redirectToCheckout({ sessionId: data.session_id });
    } else {
      throw new Error('No session ID returned');
    }
  } catch (error) {
    console.error('Checkout error:', error);
    alert('Failed to proceed to checkout. Please try again.');
    checkoutBtn.disabled = false;
    checkoutBtn.textContent = 'Proceed to Checkout';
  }
}

// ========== INITIALIZE PAGE ==========
document.addEventListener('DOMContentLoaded', async () => {
  // Initialize Stripe
  await initStripe();
  
  // Render cart
  renderCartPage();

  // Checkout button
  const checkoutBtn = document.getElementById('checkout-btn');
  if (checkoutBtn) {
    checkoutBtn.addEventListener('click', handleCheckout);
  }
});
