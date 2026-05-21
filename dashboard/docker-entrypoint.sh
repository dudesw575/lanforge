#!/bin/sh
set -e

# --- 1. OIDC Configuration ---
if [ -z "$OIDC_ISSUER" ]; then
  echo "⚠️ OIDC_ISSUER not set, defaulting to localhost Keycloak"
  OIDC_ISSUER="http://localhost/auth/realms/LanParty"
fi

echo "🔧 Injecting OIDC Authority runtime config: $OIDC_ISSUER"
find /var/www/ -name "*.js" -exec sed -i "s|VITE_APP_OIDC_ISSUER_PLACEHOLDER|$OIDC_ISSUER|g" {} +


# --- 2. WebSocket Configuration ---
if [ -z "$WS_URL" ]; then
  echo "⚠️ WS_URL not set, defaulting to localhost backend"
  WS_URL="ws://localhost:8080"
fi

echo "🔧 Injecting WebSocket runtime config: $WS_URL"
find /var/www/ -name "*.js" -exec sed -i "s|VITE_APP_WS_URL_PLACEHOLDER|$WS_URL|g" {} +


echo "🚀 Starting Nginx..."
exec "$@"