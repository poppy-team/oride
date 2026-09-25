# Example: Pull Request Security Review

This document provides a realistic example of a security review comment left on a Pull Request.

## Pull Request Context
**Title**: Add user profile image upload feature
**Code snippet under review**:
```python
@app.route('/upload_profile_picture', methods=['POST'])
@login_required
def upload():
    file = request.files['image']
    user_id = request.form['user_id']
    
    # Save file
    filename = secure_filename(file.filename)
    save_path = os.path.join(app.config['UPLOAD_FOLDER'], filename)
    file.save(save_path)
    
    # Update DB
    db.execute(f"UPDATE users SET profile_pic = '{save_path}' WHERE id = {user_id}")
    
    return "Success"
```

## Security Reviewer Comments

**Status: 🛑 Changes Requested**

Hi team, thanks for this feature. I've reviewed the code and found a few critical security issues that need to be addressed before merging:

1. **Insecure Direct Object Reference (IDOR) & Authorization Flaw**
   - *Line*: `user_id = request.form['user_id']`
   - *Issue*: The code trusts the `user_id` provided in the form POST data. Even though `@login_required` ensures the user is logged in, an attacker could change the `user_id` in the request to modify another user's profile picture.
   - *Fix*: Do not take `user_id` from the request body. Extract it from the server-side session or JWT token (e.g., `user_id = current_user.id`).

2. **SQL Injection Vulnerability**
   - *Line*: `db.execute(f"UPDATE users SET profile_pic = '{save_path}' WHERE id = {user_id}")`
   - *Issue*: The query is constructed using f-strings. If `user_id` wasn't an integer (or if the filename could somehow break out of the string), this would lead to SQL injection.
   - *Fix*: Use parameterized queries.
   - *Code suggestion*: `db.execute("UPDATE users SET profile_pic = %s WHERE id = %s", (save_path, current_user.id))`

3. **File Upload Security Flaws**
   - *Line*: `file = request.files['image']`
   - *Issue*: There is no validation on the file extension, MIME type, or file size. An attacker could upload a `.php` or `.py` web shell or a massive 10GB file to cause a DoS.
   - *Fix*: Implement strict validation. Check the MIME type (e.g., `image/jpeg`, `image/png`), verify the extension, and enforce a file size limit (e.g., 2MB max). Avoid saving files directly to the executable file system; prefer cloud storage like S3.

Please ping me once these are addressed!
