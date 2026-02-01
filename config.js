// config.js
// Environment-aware API configuration

function getApiBaseUrl() {
  const host = window.location.hostname;
  const isLocalDev = host === 'localhost' ||
                     host === '127.0.0.1' ||
                     host === '' ||
                     host.startsWith('10.') ||
                     host.startsWith('192.168.');

  if (isLocalDev) {
    return 'http://' + (host || 'localhost') + ':8080/api/v1';
  }

  // Production: API served from same domain
  const protocol = window.location.protocol;
  return `${protocol}//${window.location.host}/api/v1`;
}

// Rewrite asset URLs so images served by the backend resolve correctly
// when accessing the site from a LAN IP instead of localhost
function resolveAssetUrl(url) {
  if (!url) return url;
  var host = window.location.hostname;
  var isLocalDev = host === 'localhost' || host === '127.0.0.1' || host === '' ||
                   host.startsWith('10.') || host.startsWith('192.168.');
  if (isLocalDev && (url.indexOf('//localhost:') !== -1 || url.indexOf('//127.0.0.1:') !== -1)) {
    return url.replace(/\/\/(localhost|127\.0\.0\.1):(\d+)/, '//' + (host || 'localhost') + ':$2');
  }
  return url;
}

const API_CONFIG = {
  BASE_URL: getApiBaseUrl(),
  PRODUCTS_ENDPOINT: `${getApiBaseUrl()}/products`,
  ORDERS_ENDPOINT: `${getApiBaseUrl()}/orders`,
  CHECKOUT_ENDPOINT: `${getApiBaseUrl()}/cart/checkout`,
  CONFIG_ENDPOINT: `${getApiBaseUrl()}/config`
};
