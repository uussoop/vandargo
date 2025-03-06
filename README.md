# Vandar API Integration Guide

## Table of Contents
- [Introduction](#introduction)
- [Authentication](#authentication)
- [Base URL](#base-url)
- [Rate Limits](#rate-limits)
- [Payment Gateway (IPG)](#payment-gateway-ipg)
  - [Payment Flow](#payment-flow)
  - [Important Notes](#important-notes)
  - [1. Send Transaction Info](#1-send-transaction-info)
  - [2. Redirect to Payment Page](#2-redirect-to-payment-page)
  - [3. Get Transaction Info (Optional)](#3-get-transaction-info-optional)
  - [4. Verify Transaction](#4-verify-transaction)
- [Refund](#refund)
- [Settlement](#settlement)
- [Batch Settlement](#batch-settlement)
- [Queued Settlement](#queued-settlement)
- [Cash-in Service](#cash-in-service)
- [Direct Debit](#direct-debit)
- [Customer Management](#customer-management)
- [Error Handling](#error-handling)
- [Postman Collection](#postman-collection)

## Introduction

Vandar's APIs follow REST standards and return JSON-encoded responses. Authentication is performed by sending a token with each request (though some services may have different authentication methods).

## Authentication

Initial tokens are obtained through the Vandar dashboard with owner or manager access. Go to Settings > Tokens, click "Get New Token", provide a name and your password. You'll receive both a token and a refresh token.

Current token lifetime is 5 days, and refresh token lifetime is 10 days. You must refresh your token before expiration:

```javascript
// Refresh token example
var request = require('request')
var options = {
  method: 'POST',
  url: '/v3/refreshtoken',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    refreshtoken: '{refreshtoken}'
  })
}
request(options, function (error, response) {
  if (error) throw new Error(error)
  console.log(response.body)
})
```

## Base URL

```
https://api.vandar.io
```

## Rate Limits

- POST methods: 30 requests every 3 seconds
- Other methods: 100 requests every 3 seconds

## Payment Gateway (IPG)

The Payment Gateway allows online businesses to process payments for products or services on their website.

> **Note**: You must activate this tool in the Vandar dashboard and obtain a token.

### Payment Flow

1. Send transaction information and receive a payment token
2. Redirect the user to the payment page using the received token
3. (Optional) Retrieve transaction information before confirmation
4. After payment, the user is redirected to your callback URL. You must call the verify transaction method to finalize the transaction

### Important Notes

#### Callback URL Synchronization

The `callback_url` must match the domain registered with Shaparak. For example, if your Shaparak domain is `vandar.io`, the callback URL must be a derivative like `https://vandar.io/:path` or `https://subdomain.vandar.io/:path`.

#### HTTP Referer Synchronization

When redirecting to the payment page, the HTTP Referer header must:
- Not be empty
- Match the domain registered with Shaparak

### 1. Send Transaction Info

```javascript
// Send transaction information
var request = require('request')
var options = {
  method: 'POST',
  url: 'https://ipg.vandar.io/api/v4/send',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    api_key: 'YOUR_API_KEY',
    amount: 1000000, // Amount in Rials
    callback_url: 'https://yourdomain.com/callback',
    mobile_number: '09123456789', // Optional
    factorNumber: '123456', // Optional
    description: 'Payment for product', // Optional
    valid_card_number: '6219861012345678' // Optional
  })
}
request(options, function (error, response) {
  if (error) throw new Error(error)
  console.log(response.body)
})
```

### 2. Redirect to Payment Page

Redirect the user to the payment page using the token received in step 1:

```
https://ipg.vandar.io/v4/{token}
```

### 3. Get Transaction Info (Optional)

Before finalizing the transaction, you can retrieve information about it:

```javascript
var request = require('request')
var options = {
  method: 'POST',
  url: 'https://ipg.vandar.io/api/v4/transaction',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    api_key: 'YOUR_API_KEY',
    token: 'TRANSACTION_TOKEN'
  })
}
request(options, function (error, response) {
  if (error) throw new Error(error)
  console.log(response.body)
})
```

### 4. Verify Transaction

After the user completes the payment and is redirected to your callback URL, you must verify the transaction to finalize it:

```javascript
var request = require('request')
var options = {
  method: 'POST',
  url: 'https://ipg.vandar.io/api/v4/verify',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    api_key: 'YOUR_API_KEY',
    token: 'TRANSACTION_TOKEN'
  })
}
request(options, function (error, response) {
  if (error) throw new Error(error)
  console.log(response.body)
})
```

## Refund

You can use the refund service to return the amount of a successful payment gateway transaction to the payer's card.

**Requirements:**
- The refund tool must be activated
- The transaction must be successful
- You must have sufficient balance in your wallet

```javascript
var request = require('request')
var options = {
  method: 'POST',
  url: '/v3/business/:business/transaction/:transaction_id/refund',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer YOUR_TOKEN'
  },
  body: JSON.stringify({
    // Required parameters
  })
}
```

## Settlement

Use the settlement service to transfer an amount from your business wallet to any IBAN in the Iranian banking system.

```javascript
// Store settlement
var request = require('request')
var options = {
  method: 'POST',
  url: '/v3/business/:business/settlement/store',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer YOUR_TOKEN'
  },
  body: JSON.stringify({
    amount: 1000000, // Amount in Rials
    iban: 'IR000000000000000000000000',
    track_id: 'YOUR_TRACKING_ID', // Optional
    first_name: 'First Name', // Optional
    last_name: 'Last Name', // Optional
    description: 'Settlement description' // Optional
  })
}
```

## Batch Settlement

For batch (group) settlements:

```javascript
var request = require('request')
var options = {
  method: 'POST',
  url: 'https://batch.vandar.io/api/v2/business/:business/batches-settlement',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer YOUR_TOKEN'
  },
  body: JSON.stringify({
    // Batch settlement parameters
  })
}
```

## Queued Settlement

When your wallet doesn't have enough balance, you can queue settlements. They will be processed once your wallet balance is sufficient.

```javascript
var request = require('request')
var options = {
  method: 'POST',
  url: '/v3/business/:business/settlement/queued',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer YOUR_TOKEN'
  },
  body: JSON.stringify({
    // Queued settlement parameters
  })
}
```

## Cash-in Service

The "Cash-in" service allows you to receive payments with a unique identifier. This service is available by default for all active businesses, and each business is assigned a unique identifier.

```javascript
// Get cash-in code
var request = require('request')
var options = {
  method: 'GET',
  url: '/v3/business/:business/cash-in/code',
  headers: {
    'Authorization': 'Bearer YOUR_TOKEN'
  }
}
```

## Direct Debit

Direct debit (automatic payment) allows you to deduct amounts from a customer's bank account with their permission.

### Banks

```javascript
// Get active direct debit banks
var request = require('request')
var options = {
  method: 'GET',
  url: '/v3/business/:business/subscription/banks/actives',
  headers: {
    'Authorization': 'Bearer YOUR_TOKEN'
  }
}
```

### Payment Authorization

```javascript
// Store authorization
var request = require('request')
var options = {
  method: 'POST',
  url: '/v3/business/:business/subscription/authorization/store',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer YOUR_TOKEN'
  },
  body: JSON.stringify({
    // Authorization parameters
  })
}
```

### Account Withdrawal

```javascript
// Store withdrawal
var request = require('request')
var options = {
  method: 'POST',
  url: '/v3/business/:business/subscription/withdrawal/store',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer YOUR_TOKEN'
  },
  body: JSON.stringify({
    // Withdrawal parameters
  })
}
```

## Customer Management

Create, update, and manage customer information:

```javascript
// Get customers
var request = require('request')
var options = {
  method: 'GET',
  url: '/v2/business/:business/customers',
  headers: {
    'Authorization': 'Bearer YOUR_TOKEN'
  }
}

// Create customer
var options = {
  method: 'POST',
  url: '/v2/business/:business/customers',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer YOUR_TOKEN'
  },
  body: JSON.stringify({
    // Customer data
  })
}
```

### Customer Wallet

Manage customer wallets:

```javascript
// Get wallet balance
var request = require('request')
var options = {
  method: 'GET',
  url: '/v2/business/:business/customers/:customer/wallet',
  headers: {
    'Authorization': 'Bearer YOUR_TOKEN'
  }
}

// Deposit to wallet
var options = {
  method: 'POST',
  url: '/v2/business/:business/customers/:customer/wallet/deposit',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer YOUR_TOKEN'
  },
  body: JSON.stringify({
    amount: 1000000, // Amount in Rials
    description: 'Deposit description' // Optional
  })
}
```

## Error Handling

Vandar APIs follow REST conventions:
- 2xx: Successful responses
- 4xx: Errors due to incorrect information sent from your side
- 5xx: Errors on Vandar's side (contact support)

HTTP Status Codes:
- 200 - OK: Everything worked as expected
- 400 - Bad Request: The request was unacceptable (missing parameter)
- 401 - Unauthorized: No valid API key provided
- 402 - Request Failed: Valid parameters but request failed
- 403 - Forbidden: API key doesn't have permissions

## Postman Collection

For testing and development, use Vandar's Postman collection. It contains all Vandar services. To test APIs, replace the `business` parameter with your business's English name and the `access_token` with your token.

You can find your business's English name in the Vandar dashboard under Settings > Basic Information > Brand English Name.