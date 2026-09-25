**Vulnerable:**
`query = f"SELECT * FROM users WHERE id = {user_id}"`

**Secure:**
`cursor.execute("SELECT * FROM users WHERE id = ?", (user_id,))`
