# Before and After: Clean Code

## Before: A messy function

```python
def process_data(d):
    # check if d is valid
    if d is not None and len(d) > 0:
        res = []
        for i in d:
            if i.get('status') == 'active':
                # calculate something
                val = i.get('val1') * i.get('val2')
                if val > 100:
                    res.append({
                        'id': i.get('id'),
                        'computed': val,
                        'flag': True
                    })
                else:
                    res.append({
                        'id': i.get('id'),
                        'computed': val,
                        'flag': False
                    })
        return res
    return []
```

## After: Refactored for Clean Code

```python
def process_data(records: list[dict]) -> list[dict]:
    if not _has_valid_records(records):
        return []

    return [_process_active_record(record) for record in records if _is_active(record)]

def _has_valid_records(records: list[dict]) -> bool:
    return records is not None and len(records) > 0

def _is_active(record: dict) -> bool:
    return record.get('status') == 'active'

def _process_active_record(record: dict) -> dict:
    computed_value = _calculate_value(record)
    return {
        'id': record.get('id'),
        'computed': computed_value,
        'flag': computed_value > 100
    }

def _calculate_value(record: dict) -> float:
    return record.get('val1', 0) * record.get('val2', 0)
```
