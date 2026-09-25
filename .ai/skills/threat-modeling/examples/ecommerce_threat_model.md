# Example Threat Model: E-Commerce Payment Gateway Integration

## Component: Checkout API Service

### 1. Spoofing
- **Threat**: An attacker replays an old payment request to duplicate a transaction or process it for free.
- **Mitigation**: Use nonce (Number Used Once) for all payment API requests and ensure the payment gateway verifies idempotency keys. Require mTLS for communication between the internal checkout service and the gateway.

### 2. Tampering
- **Threat**: A user intercepts the client-to-server request and alters the product price before it hits the backend.
- **Mitigation**: The backend *never* trusts client-provided prices. The checkout service recalculates the total price based on product IDs from the secure product database before initiating the payment intent.

### 3. Repudiation
- **Threat**: A malicious user claims they never authorized a high-value purchase.
- **Mitigation**: Implement strict audit logging for all checkout actions, including IP address, user agent, timestamp, and the exact authorization token used. Store logs in an append-only, immutable data store.

### 4. Information Disclosure
- **Threat**: PCI data (credit card numbers) is accidentally logged by the application server during a crash or error.
- **Mitigation**: Implement log sanitization libraries that regex-match and mask credit card formats. Avoid passing raw PAN data through the application backend; use client-side tokenization (e.g., Stripe Elements).

### 5. Denial of Service
- **Threat**: An attacker repeatedly hits the `/checkout/initiate` endpoint to exhaust backend resources or run up external API costs.
- **Mitigation**: Implement strict IP-based and user-based rate limiting on checkout endpoints (e.g., max 5 attempts per minute). Add CAPTCHA for anomalous behavior patterns.

### 6. Elevation of Privilege
- **Threat**: A regular user accesses an administrative endpoint `/checkout/refund` by guessing the URL and bypassing authorization checks.
- **Mitigation**: Enforce Role-Based Access Control (RBAC). The application middleware must verify the JWT claims to ensure the user has the `admin:refunds` role before routing the request.
