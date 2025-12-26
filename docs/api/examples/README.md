# GIIA API Code Examples

This directory contains code examples for integrating with GIIA APIs.

## Examples

| Example | Language | Description |
|---------|----------|-------------|
| [complete_workflow.go](./complete_workflow.go) | Go | End-to-end workflow example |
| [python_client.py](./python_client.py) | Python | Full Python SDK example |
| [javascript_client.js](./javascript_client.js) | JavaScript | Browser/Node.js client |
| [curl_examples.sh](./curl_examples.sh) | cURL | Shell script examples |

## Quick Start

### Go

```go
client := giia.NewClient("http://localhost", "your_token", "org_id")
products, err := client.Catalog.ListProducts(ctx)
```

### Python

```python
from giia import GIIAClient
client = GIIAClient("http://localhost", "your_token", "org_id")
products = client.catalog.list_products()
```

### JavaScript

```javascript
import { GIIAClient } from './giia-client';
const client = new GIIAClient('http://localhost', 'your_token', 'org_id');
const products = await client.catalog.listProducts();
```

### cURL

```bash
source ./curl_examples.sh
giia_login "user@example.com" "password"
giia_list_products
```

## Running Examples

### Prerequisites

```bash
# Set environment variables
export GIIA_API_URL=http://localhost
export GIIA_EMAIL=user@example.com
export GIIA_PASSWORD=password
```

### Go

```bash
cd examples
go run complete_workflow.go
```

### Python

```bash
pip install requests
python python_client.py
```

### JavaScript

```bash
node javascript_client.js
```
