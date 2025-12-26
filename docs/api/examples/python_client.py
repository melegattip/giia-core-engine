#!/usr/bin/env python3
"""
GIIA Platform Python Client Example

This module provides a complete Python client for GIIA APIs.
Run: python python_client.py
"""

import os
import json
import requests
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Any
from dataclasses import dataclass


@dataclass
class User:
    id: str
    email: str
    first_name: str
    last_name: str
    organization_id: str
    roles: List[str]


@dataclass
class Product:
    id: str
    sku: str
    name: str
    category: str
    unit_of_measure: str
    status: str


@dataclass
class Buffer:
    product_id: str
    zone: str
    net_flow_position: float
    buffer_penetration: float
    red_zone: float
    yellow_zone: float
    green_zone: float


class GIIAClient:
    """Main client for GIIA Platform APIs."""

    def __init__(self, base_url: str = "http://localhost"):
        self.base_url = base_url
        self.access_token: Optional[str] = None
        self.org_id: Optional[str] = None
        self.user: Optional[User] = None
        self.session = requests.Session()

    @property
    def _headers(self) -> Dict[str, str]:
        """Get standard headers for API requests."""
        headers = {"Content-Type": "application/json"}
        if self.access_token:
            headers["Authorization"] = f"Bearer {self.access_token}"
        if self.org_id:
            headers["X-Organization-ID"] = self.org_id
        return headers

    # ========== Authentication ==========

    def login(self, email: str, password: str) -> User:
        """Authenticate and get access token."""
        response = self.session.post(
            f"{self.base_url}:8081/api/v1/auth/login",
            json={"email": email, "password": password}
        )
        response.raise_for_status()
        data = response.json()

        self.access_token = data["access_token"]
        self.org_id = data["user"]["organization_id"]
        self.user = User(
            id=data["user"]["id"],
            email=data["user"]["email"],
            first_name=data["user"].get("first_name", ""),
            last_name=data["user"].get("last_name", ""),
            organization_id=data["user"]["organization_id"],
            roles=data["user"].get("roles", [])
        )
        return self.user

    def refresh_token(self) -> str:
        """Refresh the access token."""
        response = self.session.post(
            f"{self.base_url}:8081/api/v1/auth/refresh"
        )
        response.raise_for_status()
        data = response.json()
        self.access_token = data["access_token"]
        return self.access_token

    def logout(self) -> None:
        """Logout and invalidate tokens."""
        self.session.post(
            f"{self.base_url}:8081/api/v1/auth/logout",
            headers=self._headers
        )
        self.access_token = None
        self.org_id = None
        self.user = None

    # ========== Products (Catalog Service) ==========

    def list_products(self, page: int = 1, page_size: int = 20, 
                      status: Optional[str] = None) -> List[Product]:
        """List all products."""
        params = {"page": page, "page_size": page_size}
        if status:
            params["status"] = status

        response = self.session.get(
            f"{self.base_url}:8082/api/v1/products",
            headers=self._headers,
            params=params
        )
        response.raise_for_status()
        data = response.json()

        return [
            Product(
                id=p["id"],
                sku=p["sku"],
                name=p["name"],
                category=p.get("category", ""),
                unit_of_measure=p.get("unit_of_measure", ""),
                status=p.get("status", "active")
            )
            for p in data.get("products", [])
        ]

    def get_product(self, product_id: str) -> Product:
        """Get a single product."""
        response = self.session.get(
            f"{self.base_url}:8082/api/v1/products/{product_id}",
            headers=self._headers
        )
        response.raise_for_status()
        p = response.json()
        return Product(
            id=p["id"],
            sku=p["sku"],
            name=p["name"],
            category=p.get("category", ""),
            unit_of_measure=p.get("unit_of_measure", ""),
            status=p.get("status", "active")
        )

    def create_product(self, sku: str, name: str, category: str = "",
                       unit_of_measure: str = "units") -> Product:
        """Create a new product."""
        response = self.session.post(
            f"{self.base_url}:8082/api/v1/products",
            headers=self._headers,
            json={
                "sku": sku,
                "name": name,
                "category": category,
                "unit_of_measure": unit_of_measure
            }
        )
        response.raise_for_status()
        p = response.json()
        return Product(
            id=p["id"],
            sku=p["sku"],
            name=p["name"],
            category=p.get("category", ""),
            unit_of_measure=p.get("unit_of_measure", ""),
            status=p.get("status", "active")
        )

    # ========== Buffers (DDMRP Engine) ==========

    def get_buffer(self, product_id: str) -> Buffer:
        """Get buffer status for a product."""
        response = self.session.get(
            f"{self.base_url}:8083/api/v1/buffers/{product_id}",
            headers=self._headers
        )
        response.raise_for_status()
        b = response.json().get("buffer", {})
        return Buffer(
            product_id=b.get("product_id", product_id),
            zone=b.get("zone", "unknown"),
            net_flow_position=b.get("net_flow_position", 0),
            buffer_penetration=b.get("buffer_penetration", 0),
            red_zone=b.get("red_zone", 0),
            yellow_zone=b.get("yellow_zone", 0),
            green_zone=b.get("green_zone", 0)
        )

    def calculate_buffer(self, product_id: str) -> Buffer:
        """Recalculate buffer for a product."""
        response = self.session.post(
            f"{self.base_url}:8083/api/v1/buffers/{product_id}/calculate",
            headers=self._headers
        )
        response.raise_for_status()
        b = response.json().get("buffer", {})
        return Buffer(
            product_id=b.get("product_id", product_id),
            zone=b.get("zone", "unknown"),
            net_flow_position=b.get("net_flow_position", 0),
            buffer_penetration=b.get("buffer_penetration", 0),
            red_zone=b.get("red_zone", 0),
            yellow_zone=b.get("yellow_zone", 0),
            green_zone=b.get("green_zone", 0)
        )

    # ========== Purchase Orders (Execution Service) ==========

    def create_purchase_order(self, po_number: str, supplier_id: str,
                               line_items: List[Dict[str, Any]],
                               expected_days: int = 14) -> Dict[str, Any]:
        """Create a purchase order."""
        response = self.session.post(
            f"{self.base_url}:8084/api/v1/purchase-orders",
            headers=self._headers,
            json={
                "po_number": po_number,
                "supplier_id": supplier_id,
                "order_date": datetime.now().strftime("%Y-%m-%d"),
                "expected_arrival_date": (datetime.now() + timedelta(days=expected_days)).strftime("%Y-%m-%d"),
                "line_items": line_items
            }
        )
        response.raise_for_status()
        return response.json()

    def list_purchase_orders(self, status: Optional[str] = None) -> List[Dict[str, Any]]:
        """List purchase orders."""
        params = {}
        if status:
            params["status"] = status

        response = self.session.get(
            f"{self.base_url}:8084/api/v1/purchase-orders",
            headers=self._headers,
            params=params
        )
        response.raise_for_status()
        return response.json().get("data", [])

    # ========== Analytics ==========

    def get_kpi_snapshot(self) -> Dict[str, Any]:
        """Get current KPI snapshot."""
        response = self.session.get(
            f"{self.base_url}:8085/api/v1/analytics/snapshot",
            headers=self._headers
        )
        response.raise_for_status()
        return response.json()

    # ========== Notifications (AI Hub) ==========

    def get_notifications(self, unread_only: bool = False) -> List[Dict[str, Any]]:
        """Get user notifications."""
        params = {"unread_only": str(unread_only).lower()}
        response = self.session.get(
            f"{self.base_url}:8086/api/v1/notifications",
            headers=self._headers,
            params=params
        )
        response.raise_for_status()
        return response.json().get("notifications", [])


# ========== Example Usage ==========

def main():
    """Demonstrate GIIA client usage."""
    print("=" * 50)
    print("GIIA Platform Python Client Example")
    print("=" * 50)

    # Initialize client
    client = GIIAClient(os.environ.get("GIIA_API_URL", "http://localhost"))

    # Step 1: Login
    print("\n1. Authenticating...")
    email = os.environ.get("GIIA_EMAIL", "demo@example.com")
    password = os.environ.get("GIIA_PASSWORD", "password")

    try:
        user = client.login(email, password)
        print(f"   ✓ Logged in as {user.email}")
        print(f"   ✓ Organization: {user.organization_id}")
    except requests.exceptions.RequestException as e:
        print(f"   ✗ Login failed: {e}")
        return

    # Step 2: List products
    print("\n2. Listing products...")
    try:
        products = client.list_products(page_size=5)
        print(f"   ✓ Found {len(products)} products")
        for p in products[:3]:
            print(f"      - {p.sku}: {p.name}")
    except requests.exceptions.RequestException as e:
        print(f"   ✗ Failed: {e}")
        products = []

    # Step 3: Create a product
    print("\n3. Creating product...")
    try:
        new_product = client.create_product(
            sku=f"PY-{datetime.now().strftime('%Y%m%d%H%M%S')}",
            name="Python Demo Product",
            category="Demo",
            unit_of_measure="units"
        )
        print(f"   ✓ Created: {new_product.sku} - {new_product.name}")
    except requests.exceptions.RequestException as e:
        print(f"   ✗ Failed: {e}")

    # Step 4: Check buffer
    print("\n4. Checking buffer status...")
    if products:
        try:
            buffer = client.get_buffer(products[0].id)
            print(f"   ✓ Buffer for {products[0].sku}:")
            print(f"      Zone: {buffer.zone}")
            print(f"      NFP: {buffer.net_flow_position:.2f}")
            print(f"      Penetration: {buffer.buffer_penetration * 100:.1f}%")
        except requests.exceptions.RequestException as e:
            print(f"   ✗ Failed: {e}")

    # Step 5: Get KPI snapshot
    print("\n5. Getting KPI snapshot...")
    try:
        kpis = client.get_kpi_snapshot()
        print(f"   ✓ Service Level: {kpis.get('service_level', 0) * 100:.1f}%")
        print(f"   ✓ Stockout Rate: {kpis.get('stockout_rate', 0) * 100:.1f}%")
    except requests.exceptions.RequestException as e:
        print(f"   ✗ Failed: {e}")

    # Step 6: Get notifications
    print("\n6. Getting notifications...")
    try:
        notifications = client.get_notifications(unread_only=True)
        print(f"   ✓ {len(notifications)} unread notifications")
        for n in notifications[:3]:
            print(f"      - [{n.get('priority', 'info')}] {n.get('title', 'Untitled')}")
    except requests.exceptions.RequestException as e:
        print(f"   ✗ Failed: {e}")

    print("\n" + "=" * 50)
    print("Example complete!")
    print("=" * 50)


if __name__ == "__main__":
    main()
