/**
 * config.js
 * Centralized configuration for Nessie Audio website
 * Automatically detects environment (local development vs production)
 */

/**
 * Determine the API base URL based on the current environment
 * @returns {string} The appropriate API base URL
 */
function getApiBaseUrl() {
  // Check if we're running on localhost (development)
  const isLocalhost = window.location.hostname === 'localhost' ||
                      window.location.hostname === '127.0.0.1' ||
                      window.location.hostname === '';

  if (isLocalhost) {
    // Local development - use localhost backend
    return 'http://localhost:8080/api/v1';
  } else {
    // Production - use the same domain as the frontend
    // This assumes your backend API is served from the same domain
    // e.g., https://nessieaudio.com/api/v1
    const protocol = window.location.protocol; // http: or https:
    const host = window.location.host; // domain.com or domain.com:port
    return `${protocol}//${host}/api/v1`;

    // Alternative: If your API is on a subdomain, uncomment and modify this:
    // return `${protocol}//api.${window.location.hostname}/api/v1`;
  }
}

// Export configuration
const API_CONFIG = {
  BASE_URL: getApiBaseUrl(),
  PRODUCTS_ENDPOINT: `${getApiBaseUrl()}/products`,
  ORDERS_ENDPOINT: `${getApiBaseUrl()}/orders`,
  CHECKOUT_ENDPOINT: `${getApiBaseUrl()}/cart/checkout`,
  CONFIG_ENDPOINT: `${getApiBaseUrl()}/config`
};

// Log the current API URL for debugging (remove in production if desired)
console.log('API Configuration loaded:', API_CONFIG.BASE_URL);
