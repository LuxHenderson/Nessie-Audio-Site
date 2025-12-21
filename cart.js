/**
 * cart.js
 * Shopping cart management for Nessie Audio merch store
 * Handles cart state, localStorage persistence, and cart operations
 */

// ========== CART STATE MANAGEMENT ==========
class ShoppingCart {
  constructor() {
    this.storageKey = 'nessie_audio_cart';
    this.items = this.loadCart();
  }

  /**
   * Load cart from localStorage
   * @returns {Array} Cart items
   */
  loadCart() {
    try {
      const saved = localStorage.getItem(this.storageKey);
      return saved ? JSON.parse(saved) : [];
    } catch (error) {
      console.error('Error loading cart:', error);
      return [];
    }
  }

  /**
   * Save cart to localStorage
   */
  saveCart() {
    try {
      localStorage.setItem(this.storageKey, JSON.stringify(this.items));
      this.updateCartUI();
    } catch (error) {
      console.error('Error saving cart:', error);
    }
  }

  /**
   * Add item to cart
   * @param {Object} item - Item object with productId, variantId, productName, variantName, price, image, quantity
   */
  addItem(item) {
    // Check if item already exists in cart
    const existingItem = this.items.find(
      i => i.productId === item.productId && i.variantId === item.variantId
    );

    if (existingItem) {
      existingItem.quantity += item.quantity || 1;
    } else {
      this.items.push({
        productId: item.productId,
        variantId: item.variantId,
        name: item.productName,
        variantName: item.variantName,
        price: item.price,
        imageUrl: item.image,
        quantity: item.quantity || 1
      });
    }

    this.saveCart();
    this.updateCartUI();
    // Notification is handled by product-detail.js
  }

  /**
   * Update item quantity
   * @param {string} productId - Product ID
   * @param {string} variantId - Variant ID
   * @param {number} quantity - New quantity
   */
  updateQuantity(productId, variantId, quantity) {
    const item = this.items.find(
      item => item.productId === productId && item.variantId === variantId
    );

    if (item) {
      if (quantity <= 0) {
        this.removeItem(productId, variantId);
      } else {
        item.quantity = quantity;
        this.saveCart();
      }
    }
  }

  /**
   * Remove item from cart
   * @param {string} productId - Product ID
   * @param {string} variantId - Variant ID
   */
  removeItem(productId, variantId) {
    this.items = this.items.filter(
      item => !(item.productId === productId && item.variantId === variantId)
    );
    this.saveCart();
  }

  /**
   * Clear entire cart
   */
  clearCart() {
    this.items = [];
    this.saveCart();
  }

  /**
   * Get total item count
   * @returns {number} Total number of items
   */
  getItemCount() {
    return this.items.reduce((total, item) => total + item.quantity, 0);
  }

  /**
   * Get cart subtotal
   * @returns {number} Subtotal amount
   */
  getSubtotal() {
    return this.items.reduce((total, item) => total + (item.price * item.quantity), 0);
  }

  /**
   * Get all cart items
   * @returns {Array} Cart items
   */
  getItems() {
    return this.items;
  }

  /**
   * Update cart UI elements (badge count)
   */
  updateCartUI() {
    const cartCount = this.getItemCount();
    const badges = document.querySelectorAll('.cart-count');
    
    badges.forEach(badge => {
      badge.textContent = cartCount;
      badge.style.display = cartCount > 0 ? 'inline-block' : 'none';
    });
  }

  /**
   * Show notification when item added
   * @param {string} productName - Name of added product
   */
  showAddedNotification(productName) {
    // Create notification element
    const notification = document.createElement('div');
    notification.className = 'cart-notification';
    notification.style.cssText = `
      position: fixed;
      top: 180px;
      right: 20px;
      z-index: 2001;
    `;
    notification.innerHTML = `
      <p>✓ ${productName} added to cart</p>
      <a href="cart.html">View Cart</a>
    `;
    
    document.body.appendChild(notification);

    // Animate in
    setTimeout(() => notification.classList.add('show'), 10);

    // Remove after 3 seconds
    setTimeout(() => {
      notification.classList.remove('show');
      setTimeout(() => notification.remove(), 300);
    }, 3000);
  }
}

// ========== INITIALIZE GLOBAL CART ==========
const cart = new ShoppingCart();
window.cart = cart; // Make globally accessible

// Update cart UI on page load
document.addEventListener('DOMContentLoaded', () => {
  cart.updateCartUI();
});

// Export for use in other modules (if using modules)
if (typeof module !== 'undefined' && module.exports) {
  module.exports = { cart, ShoppingCart };
}
