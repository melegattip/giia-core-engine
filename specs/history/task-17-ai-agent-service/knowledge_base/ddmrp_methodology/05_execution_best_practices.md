# DDMRP Execution Best Practices

## Overview

Execution is where DDMRP theory meets reality. Perfect buffer calculations mean nothing if execution is poor. This document outlines best practices for order generation, supplier management, and exception handling in a demand-driven environment.

## Order Generation Principles

### Pull-Based Ordering
DDMRP uses actual buffer status to drive replenishment, not forecasts or scheduled orders.

**Core Principle:**
```
Order when buffer penetrates into yellow zone
Order quantity brings inventory back to top of buffer
```

### Order Spike Generation

**Standard Order:**
```
Trigger: On-hand + On-order ≤ Top of Yellow (Order Point)
Quantity: Green Zone
Frequency: As needed (demand-driven)

Example:
Buffer Profile:
- Total Buffer: 300 units
- Green Zone: 80 units
- Yellow Zone: 100 units
- Red Zone: 120 units
- Top of Yellow: 220 units

Current Status:
- On-hand: 210 units
- On-order: 0 units
- Net Position: 210 units

Action: No order (above order point)

Next Day:
- On-hand: 190 units (20 units consumed)
- On-order: 0 units
- Net Position: 190 units

Action: Generate order spike!
Order Quantity: 80 units (Green Zone)
Expected Result: 190 + 80 = 270 units when order arrives
```

**Spike Order Timing:**
The key to DDMRP execution is frequency. Check buffer status:
- **Critical items (A):** Daily
- **Important items (B):** 2-3 times per week
- **Regular items (C):** Weekly

### Qualified Demand Concept

Not all demand should trigger immediate buffer penetration.

**Qualified vs. Unqualified Demand:**

**Qualified Demand:** Normal, expected demand that consumes buffer
- Regular customer orders
- Production consumption
- Distribution transfers

**Unqualified Demand:** Abnormal demand that should not deplete buffer
- One-time large projects
- Promotional campaigns (planned)
- Emergency customer orders (unexpected)

**Handling Unqualified Demand:**
```
Option 1: Bypass Buffer
- Source directly from supplier
- Don't let it penetrate buffer
- Maintain buffer for qualified demand

Option 2: Temporary Buffer Increase
- Increase buffer temporarily for promotional period
- Return to normal buffer after event
- Example: Christmas season, Black Friday

Example - Promotional Campaign:
Normal ADU: 10 units/day
Promotional period: 60 units/day for 14 days
Expected promotion demand: 60 × 14 = 840 units

Buffer Approach:
1. Don't deplete existing buffer (300 units)
2. Place separate promotional order: 840 units
3. Keep buffer intact for post-promotion demand
```

## Supplier Management

### Lead Time Management

**Decoupled Lead Time (DLT):**
The cumulative lead time from your decoupling point to final delivery.

```
DLT = Supplier Lead Time + Internal Processing Time + Transit Time + Receiving Time

Example:
- Supplier processing: 7 days
- Manufacturing time: 3 days
- Shipping transit: 2 days
- Receiving inspection: 1 day
Total DLT = 13 days

Buffer Red Zone = ADU × DLT
If ADU = 10 units/day
Red Zone = 10 × 13 = 130 units
```

**Lead Time Variability:**
Track actual lead times, not quoted times.

```
Metric: Lead Time Coefficient of Variation (CV)

CV = Standard Deviation of Lead Times / Average Lead Time

Example:
Supplier A: Average LT = 10 days, Std Dev = 1 day, CV = 0.10 (very reliable)
Supplier B: Average LT = 10 days, Std Dev = 3 days, CV = 0.30 (unreliable)

Buffer Implication:
Supplier A: Use VF = 0.50 (standard)
Supplier B: Use VF = 0.75 (increased protection against variability)
```

### Supplier Selection Criteria

When choosing between suppliers for emergency orders or regular replenishment:

**Priority Factors:**

1. **Lead Time (Highest Priority for Stockouts)**
```
Scenario: Red zone penetration, 2 days until stockout
Revenue at risk: $15,000

Supplier Options:
A: 7-day lead time, $10/unit
B: 2-day lead time, $13/unit (+30% premium)
C: 1-day lead time, $15/unit (+50% premium)

Analysis:
Option A: Too slow - stockout occurs, lose $15,000
Option B: ✓ Prevents stockout, premium cost $3/unit × 50 units = $150
Option C: Overkill - extra $5/unit × 50 units = $250 vs. Option B

Decision: Supplier B (fastest sufficient option)
```

2. **Reliability (On-Time Delivery)**
```
Track supplier OTD (On-Time Delivery) percentage:

Grade A: 95-100% OTD → Standard buffer factors
Grade B: 85-94% OTD → Increase VF by 10-20%
Grade C: 75-84% OTD → Increase VF by 20-30%
Grade D: <75% OTD → Consider alternative supplier

Example:
Primary supplier OTD dropped from 98% to 82% over 3 months

Action:
1. Immediate: Increase buffer VF from 0.60 to 0.75 (+25%)
2. Short-term: Dual-source critical items
3. Long-term: Supplier development program or switch
```

3. **Total Cost (When Time Permits)**
```
Total Cost of Ownership = Unit Price + Freight + Handling + Carrying Cost

Scenario: Normal replenishment (not emergency)
Order Quantity: 100 units
Annual Volume: 1,200 units

Supplier A:
- Unit Price: $10.00
- Freight: $50 flat
- Lead Time: 14 days
- Buffer required: 200 units

Supplier B:
- Unit Price: $9.50
- Freight: $80 flat
- Lead Time: 21 days
- Buffer required: 260 units (larger due to longer LT)

Annual TCO Comparison:
Supplier A:
- Product: 1,200 × $10.00 = $12,000
- Freight: 12 orders × $50 = $600
- Carrying Cost: 200 units × $10 × 25% = $500
- Total: $13,100

Supplier B:
- Product: 1,200 × $9.50 = $11,400 (save $600)
- Freight: 12 orders × $80 = $960 (+$360)
- Carrying Cost: 260 units × $9.50 × 25% = $617.50 (+$117.50)
- Total: $12,977.50

Savings with Supplier B: $122.50/year

Decision: Supplier B slightly better TCO, IF lead time reliability is acceptable
```

### Minimum Order Quantities (MOQ)

MOQs can conflict with DDMRP spike orders. Strategies to handle:

**Strategy 1: Negotiate MOQ Reduction**
```
Show supplier the benefit of more frequent, smaller orders:
- Better cash flow for you
- More predictable demand for them
- Lower inventory risk
```

**Strategy 2: Adjust Order Point**
```
If MOQ > Green Zone, adjust ordering logic:

Example:
Green Zone: 50 units (ideal order quantity)
Supplier MOQ: 100 units

Option A: Order MOQ when hitting yellow zone
- Orders less frequent
- Higher peak inventory
- Still better than traditional MRP

Option B: Share MOQ across multiple products
- Combine orders from same supplier
- Split MOQ efficiently
```

**Strategy 3: Multiple Suppliers**
```
For high-volume items where MOQ is constraining:
- Primary supplier: 70% volume, flexible small orders
- Secondary supplier: 30% volume, bulk orders with MOQ
- Balance cost vs. flexibility
```

## Exception Handling

### Stockout Prevention

**Early Warning System:**
```
Alert Levels based on Buffer Penetration:

Yellow Zone Entry (48% penetration):
- Priority: Medium
- Action: Monitor daily
- Notification: Planner

Lower Yellow (60% penetration):
- Priority: High
- Action: Verify order status
- Notification: Planner + Supervisor

Red Zone Entry (48% remaining):
- Priority: Critical
- Action: Expedite existing order OR place emergency order
- Notification: Planner + Supervisor + Procurement Manager
```

**Stockout Imminent (72 hours or less):**

**Step 1: Assess Situation**
```
Calculate:
- Days until stockout = On-hand / ADU
- Days until next receipt = Earliest order arrival
- Stockout gap = Days until stockout - Days until receipt

Example:
On-hand: 30 units
ADU: 20 units/day
Days to stockout: 30/20 = 1.5 days

Existing order: 100 units, arriving in 5 days
Stockout gap: 1.5 - 5 = -3.5 days (will stockout 3.5 days before order arrives!)

Revenue at risk: ADU × Stockout days × Unit price
Revenue at risk: 20 × 3.5 × $50 = $3,500
```

**Step 2: Immediate Actions (Priority Order)**

1. **Check for Available Inventory**
```
- Other warehouses/locations
- Work-in-process that can be expedited
- Customer returns/refurbished stock
- Similar substitute products
```

2. **Expedite Existing Order**
```
Contact supplier:
- Can they ship partial order sooner?
- Expedited shipping (air vs. ocean)
- Premium freight cost vs. stockout cost

Example:
Existing order: 100 units, standard shipping, 5 days
Option: Ship 50 units via air freight, arrive 2 days
Cost: $200 premium freight
Benefit: Prevents stockout, saves $3,500 revenue
Decision: ✓ Expedite
```

3. **Emergency Order from Faster Supplier**
```
Alternative supplier with 2-day lead time
Price premium: 25%
Order quantity: 70 units (cover gap)

Cost: 70 × $50 × 1.25 = $4,375
Standard cost would be: 70 × $50 = $3,500
Premium paid: $875

Stockout cost avoided: $3,500+
Decision: ✓ Place emergency order
```

4. **Customer Communication (Last Resort)**
```
If stockout unavoidable:
- Proactive communication
- Offer alternatives (substitute products, partial shipment)
- Expedited delivery when available
- Discount/goodwill gesture

Better to manage expectations than surprise stockout
```

### Supplier Delivery Failures

**Order Not Delivered On-Time:**

**Immediate Response:**
```
1. Contact supplier for status (within 24 hours of late)
2. Get revised delivery commitment
3. Assess impact on buffer status
4. Determine if emergency action needed

Example:
Order expected: Day 14
Current day: Day 15 (1 day late)
Current on-hand: 45 units
ADU: 10 units/day
Red zone threshold: 120 units

Status: Already in red zone (45 < 120)
Days until stockout: 45/10 = 4.5 days

Action:
- Contact supplier immediately
- If delay >2 more days, place emergency order elsewhere
- Document supplier performance issue
```

**Pattern of Late Deliveries:**
```
Track supplier On-Time Delivery (OTD) weekly:

Month 1: 95% OTD → Acceptable
Month 2: 89% OTD → Warning sign
Month 3: 78% OTD → Unacceptable

Response to Declining Performance:
1. Immediate: Increase buffer by 15-25% to protect against unreliability
2. Short-term: Dual-source critical items
3. Medium-term: Supplier improvement program
4. Long-term: If no improvement, transition to alternative supplier
```

### Demand Spikes

**Unexpected Demand Surge:**

**Scenario:**
```
Normal ADU: 10 units/day
Sudden spike: 60 units ordered in 1 day
Buffer status: Dropped from green to deep red

Question: Is this new normal or one-time spike?
```

**Assessment Process:**
```
1. Investigate Cause:
   - New customer (ongoing)?
   - Promotional campaign (temporary)?
   - Competitor stockout (temporary)?
   - Market trend change (ongoing)?

2. Response Based on Cause:

   One-Time Spike:
   - Place one-time replenishment order
   - Do NOT adjust buffer (would create excess inventory)
   - Monitor for recurrence

   Sustained Increase:
   - Immediate large order to restore buffer
   - Increase ADU calculation
   - Recalculate buffer with new ADU
   - Adjust order points

Example - Sustained Increase:
Week 1: Average daily demand = 15 (up from 10)
Week 2: Average daily demand = 18
Week 3: Average daily demand = 16

New ADU = (15 + 18 + 16) / 3 = 16.3 ≈ 16 units/day

Recalculate Buffer:
Old buffer (ADU 10, LT 12 days):
- Red: 120, Yellow: 72, Green: 60, Total: 252

New buffer (ADU 16, LT 12 days):
- Red: 192, Yellow: 115, Green: 96, Total: 403

Immediate order needed: 403 - 252 = 151 units (beyond normal spike)
```

### Execution Failure Patterns

**Multiple Failures Same Supplier:**

Detect patterns across execution history:

**Pattern Detection Rules:**
```
Rule: 3+ execution failures in 6 hours from same supplier

Example Log:
Friday 3:15 PM: Order #1234 to GlobalParts - Execution Failed
Friday 4:20 PM: Order #1235 to GlobalParts - Execution Failed
Friday 5:45 PM: Order #1236 to GlobalParts - Execution Failed

Root Cause Analysis:
- All failures Friday afternoon
- Supplier system locks inventory 5 PM Friday
- No weekend processing
- Orders fail until Monday 8 AM

Impact:
- 5 orders pending (Friday + weekend)
- $12,000 in pending orders at risk
- Monday deliveries delayed 2-3 days
- Potential stockouts on 3 products
```

**Automated Response:**
```
1. Alert: Pattern detected notification
   - Planner notified immediately
   - Escalation to procurement manager

2. Impact Assessment:
   - Which products affected
   - Buffer status of affected products
   - Revenue at risk calculation

3. Recommended Actions:
   - Place Friday orders before 3 PM deadline
   - Adjust buffer for this supplier (+48 hours to account for weekend gap)
   - Route urgent Friday afternoon orders to alternative supplier

4. Long-Term Solution:
   - Supplier integration improvement
   - Automated order timing rules
   - Alternative supplier development for backup
```

## Performance Metrics

### Execution Performance Tracking

**Key Metrics:**

1. **Service Level**
```
Service Level = (Orders Shipped On-Time / Total Orders) × 100%

Target: 95%+ on-time delivery

Example:
Monthly orders: 1,000
Shipped on-time: 970
Service level: 97% ✓
```

2. **Inventory Turns**
```
Inventory Turns = Cost of Goods Sold / Average Inventory Value

Target: Increase turns while maintaining service level

Example:
Annual COGS: $2,400,000
Average inventory: $400,000
Turns = 2,400,000 / 400,000 = 6 turns/year

With DDMRP optimization:
Reduced average inventory to $300,000
New turns = 2,400,000 / 300,000 = 8 turns/year (+33% improvement)
```

3. **Order Spike Frequency**
```
Spike Frequency = Number of Orders Generated / Time Period

High-frequency ordering = responsive demand-driven system

Pre-DDMRP (MRP weekly batches):
- 52 orders/year
- Average order size: 500 units

Post-DDMRP (demand-driven spikes):
- 156 orders/year (3× frequency)
- Average order size: 180 units
- Lower peak inventory, higher responsiveness
```

4. **Buffer Performance**
```
Buffer Penetration Distribution:
- % time in Green Zone (target: 40-50%)
- % time in Yellow Zone (target: 30-40%)
- % time in Red Zone (target: 10-20%)
- % time stocked out (target: <2%)

Example - Well-Performing Buffer:
Green: 45% of time ✓
Yellow: 35% of time ✓
Red: 18% of time ✓
Stockout: 2% of time ✓

Example - Under-Buffered Item:
Green: 15% of time ✗ (too low)
Yellow: 25% of time
Red: 50% of time ✗ (too high)
Stockout: 10% of time ✗ (unacceptable)

Action: Increase buffer by 30-40%
```

## Best Practices Summary

### DO's ✅

1. **Order Frequently in Small Batches**
   - Reduces inventory
   - Increases responsiveness
   - Minimizes obsolescence risk

2. **Monitor Buffers Daily (Critical Items)**
   - Proactive vs. reactive
   - Prevent stockouts before they occur

3. **Track Actual Lead Times**
   - Use real data, not supplier quotes
   - Adjust buffers based on actual performance

4. **Expedite When Justified**
   - Calculate revenue at risk vs. expedite cost
   - Premium freight to prevent stockout often pays for itself

5. **Communicate with Suppliers**
   - Share demand signals
   - Collaborate on lead time reduction
   - Develop partnership relationship

6. **Document Exceptions**
   - Build historical knowledge
   - Identify recurring patterns
   - Support continuous improvement

### DON'Ts ❌

1. **Don't Wait for Stockout to Act**
   - Red zone = urgent action
   - Don't wait until zero inventory

2. **Don't Ignore Demand Spikes**
   - Investigate cause immediately
   - Adjust buffers if sustained increase

3. **Don't Over-Order in Panic**
   - Order what's needed (Green Zone typically)
   - Don't create excess inventory in overreaction

4. **Don't Ignore Supplier Performance Trends**
   - Declining OTD requires action
   - Don't hope it will improve without intervention

5. **Don't Batch Orders Unnecessarily**
   - DDMRP benefits from frequent small orders
   - Batching defeats demand-driven purpose

6. **Don't Set and Forget Buffers**
   - Regular review and adjustment critical
   - Quarterly minimum, monthly for critical items

## Execution Decision Tree

```
Daily Buffer Status Check
        ↓
    Is position ≤ Top of Yellow?
        ↓
    YES → Generate Order Spike
            ↓
        Check MOQ
            ↓
        Order Quantity = max(Green Zone, MOQ)
            ↓
        Is position in Red Zone?
            ↓
        YES → Expedite delivery?
                ↓
            Revenue at risk vs. expedite cost
                ↓
            If ROI positive → Expedite
                ↓
        NO → Standard delivery
            ↓
    Place Order
        ↓
    Monitor order until delivery
        ↓
    Track actual lead time
        ↓
    Update buffer performance metrics
```

## Conclusion

Effective DDMRP execution requires:
- **Discipline:** Check buffers regularly, act on signals
- **Speed:** Respond quickly when buffers penetrate yellow/red
- **Judgment:** Know when to expedite, when to use alternative suppliers
- **Communication:** Keep suppliers informed, maintain relationships
- **Continuous Improvement:** Track metrics, learn from exceptions, optimize buffers

The goal is not perfection, but continuous improvement in service level while reducing inventory investment. Every exception is a learning opportunity to refine buffers and execution processes.