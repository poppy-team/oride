```python
def calculate_monthly_revenue(invoices):
    return sum(invoice.amount for invoice in invoices if invoice.is_paid)
```
