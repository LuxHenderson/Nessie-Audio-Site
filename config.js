// config.js
// Environment-aware API configuration

function getApiBaseUrl() {
  const isLocalhost = window.location.hostname === 'localhost' ||
                      window.location.hostname === '127.0.0.1' ||
                      window.location.hostname === '';

  if (isLocalhost) {
    return 'http://localhost:8080/api/v1';
  }

  // Production: API served from same domain
  const protocol = window.location.protocol;
  const host = window.location.host;
  return `${protocol}//${host}/api/v1`;
}

const API_CONFIG = {
  BASE_URL: getApiBaseUrl(),
  PRODUCTS_ENDPOINT: `${getApiBaseUrl()}/products`,
  ORDERS_ENDPOINT: `${getApiBaseUrl()}/orders`,
  CHECKOUT_ENDPOINT: `${getApiBaseUrl()}/cart/checkout`,
  CONFIG_ENDPOINT: `${getApiBaseUrl()}/config`
};
