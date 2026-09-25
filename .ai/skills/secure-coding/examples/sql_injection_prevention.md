# Secure Coding Example: Preventing SQL Injection

SQL Injection (SQLi) occurs when untrusted user input is directly concatenated into a database query. This allows attackers to manipulate the query logic.

## ❌ Vulnerable Example (Python/psycopg2)

In this example, string formatting is used to build the query, which is extremely dangerous.

```python
def get_user(cursor, username):
    # BAD: Using f-strings or string concatenation
    query = f"SELECT id, email FROM users WHERE username = '{username}'"
    cursor.execute(query)
    return cursor.fetchone()

# An attacker could pass: username = "admin' OR '1'='1"
# Resulting query: SELECT id, email FROM users WHERE username = 'admin' OR '1'='1'
# This bypasses authentication or extracts all users.
```

## ✅ Secure Example (Python/psycopg2)

The correct approach is to use parameterized queries (prepared statements). The database driver safely handles escaping the parameters.

```python
def get_user_secure(cursor, username):
    # GOOD: Using parameterized queries
    query = "SELECT id, email FROM users WHERE username = %s"
    
    # Pass the parameters as a separate argument (tuple) to execute()
    cursor.execute(query, (username,))
    return cursor.fetchone()

# Even if an attacker passes "admin' OR '1'='1", it is treated strictly as a string literal.
# Resulting logic: Search for a user whose literal username string is "admin' OR '1'='1"
```

## ✅ Secure Example (Node.js/pg)

Similar principles apply across languages and frameworks.

```javascript
// Vulnerable
const query = `SELECT id FROM users WHERE email = '${req.body.email}'`;
await client.query(query);

// Secure
const query = 'SELECT id FROM users WHERE email = $1';
const values = [req.body.email];
await client.query(query, values);
```
