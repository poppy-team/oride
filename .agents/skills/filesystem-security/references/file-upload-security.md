# File Upload Security

Allowing users to upload files is inherently dangerous. A malicious upload can lead to Remote Code Execution (RCE), Cross-Site Scripting (XSS), or Denial of Service (DoS).

## Security Controls

### 1. File Type Validation (Allowlist)
- **Do not rely on the file extension.** Attackers can rename `shell.php` to `shell.jpg`.
- **Do not rely on the `Content-Type` header.** This header is provided by the client and is easily spoofed.
- **Do:** Verify the file signature (magic numbers). Use a robust library to determine the true MIME type based on file content.

### 2. Prevent Execution
- **Never store uploaded files in the webroot** (e.g., `/var/www/html/uploads`). If you do, a user could upload `shell.php` and then navigate to `https://yoursite.com/uploads/shell.php` to execute it.
- Store files outside the web root or on a separate cloud storage service (e.g., AWS S3, Google Cloud Storage).
- If serving files locally, configure the web server (Nginx/Apache) to deny execution of scripts in the upload directory.

### 3. Rename Files
- **Never trust the user-provided filename.** It could contain path traversal characters (e.g., `../../../etc/passwd`) or malicious characters.
- Generate a new, random filename (e.g., UUID) for every upload. Store the original filename in the database if necessary.

### 4. Limit File Size
- Enforce strict file size limits to prevent Denial of Service (disk exhaustion).
- Reject files that exceed the limit at the web server/proxy level (e.g., Nginx `client_max_body_size`) before they hit your application logic.

### 5. Virus Scanning
- Integrate a virus scanner (like ClamAV) to scan all uploaded files before storing them permanently.

### 6. Image Stripping
- If accepting images, strip EXIF data. EXIF data can contain sensitive information (GPS coordinates) or malicious payloads.
