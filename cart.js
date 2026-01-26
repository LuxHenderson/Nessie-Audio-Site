// cart.js
// Shopping cart with localStorage persistence

class ShoppingCart {
  constructor() {
    this.storageKey = 'nessie_audio_cart';
    this.items = this.loadCart();
  }

  loadCart() {
    try {
      const saved = localStorage.getItem(this.storageKey);
      return saved ? JSON.parse(saved) : [];
    } catch (error) {
      console.error('Error loading cart:', error);
      return [];
    }
  }

  saveCart() {
    try {
      localStorage.setItem(this.storageKey, JSON.stringify(this.items));
      this.updateCartUI();
    } catch (error) {
      console.error('Error saving cart:', error);
    }
  }

  addItem(item) {
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
  }

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

  removeItem(productId, variantId) {
    this.items = this.items.filter(
      item => !(item.productId === productId && item.variantId === variantId)
    );
    this.saveCart();
  }

  clearCart() {
    this.items = [];
    this.saveCart();
  }

  getItemCount() {
    return this.items.reduce((total, item) => total + item.quantity, 0);
  }

  getSubtotal() {
    return this.items.reduce((total, item) => total + (item.price * item.quantity), 0);
  }

  getItems() {
    return this.items;
  }

  updateCartUI() {
    const cartCount = this.getItemCount();
    const badges = document.querySelectorAll('.cart-count');

    badges.forEach(badge => {
      badge.textContent = cartCount;
      badge.style.display = cartCount > 0 ? 'inline-block' : 'none';
      badge.setAttribute('aria-label', `${cartCount} ${cartCount === 1 ? 'item' : 'items'} in cart`);
    });
  }

  announceToScreenReader(message) {
    let announcement = document.getElementById('sr-announcement');

    if (!announcement) {
      announcement = document.createElement('div');
      announcement.id = 'sr-announcement';
      announcement.setAttribute('role', 'status');
      announcement.setAttribute('aria-live', 'polite');
      announcement.setAttribute('aria-atomic', 'true');
      announcement.className = 'sr-only';
      document.body.appendChild(announcement);
    }

    announcement.textContent = message;

    // Clear to allow subsequent announcements to trigger
    setTimeout(() => {
      announcement.textContent = '';
    }, 1000);
  }

  showAddedNotification(productName) {
    this.announceToScreenReader(`${productName} added to cart. ${this.getItemCount()} items in cart.`);

    const notification = document.createElement('div');
    notification.className = 'cart-notification';
    notification.setAttribute('role', 'alert');
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

    setTimeout(() => notification.classList.add('show'), 10);

    setTimeout(() => {
      notification.classList.remove('show');
      setTimeout(() => notification.remove(), 300);
    }, 3000);
  }
}

const cart = new ShoppingCart();
window.cart = cart;

document.addEventListener('DOMContentLoaded', () => {
  cart.updateCartUI();
});

if (typeof module !== 'undefined' && module.exports) {
  module.exports = { cart, ShoppingCart };
}
