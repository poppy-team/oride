# Example: Secure File Upload (Python/Flask)

This example demonstrates secure file upload practices: using magic numbers for MIME typing, UUIDs for filenames, and saving outside the webroot.

```python
import os
import uuid
import magic # python-magic library for MIME type detection
from werkzeug.utils import secure_filename
from flask import Flask, request, abort

app = Flask(__name__)
# Keep uploads outside of the static folder!
UPLOAD_FOLDER = '/var/app/data/uploads'
# Strict allowlist of MIME types
ALLOWED_MIMETYPES = {'image/jpeg', 'image/png'}
MAX_FILE_SIZE = 2 * 1024 * 1024 # 2 MB

@app.route('/upload', methods=['POST'])
def upload_file():
    if 'file' not in request.files:
        return abort(400, "No file part")
    
    file = request.files['file']
    if file.filename == '':
        return abort(400, "No selected file")

    # 1. Read a chunk to determine the true MIME type
    file_bytes = file.read(2048)
    mime_type = magic.from_buffer(file_bytes, mime=True)
    
    # Reset file pointer after reading
    file.seek(0, os.SEEK_END)
    file_size = file.tell()
    file.seek(0)
    
    # 2. Enforce File Size Limit
    if file_size > MAX_FILE_SIZE:
        return abort(413, "File too large")

    # 3. Enforce strict MIME type allowlist
    if mime_type not in ALLOWED_MIMETYPES:
        return abort(415, "Unsupported Media Type")

    # 4. Generate a completely safe, random filename. 
    # Extract the extension from the original filename (validated by werkzeug)
    original_filename = secure_filename(file.filename)
    ext = os.path.splitext(original_filename)[1].lower()
    
    # We only append the extension if it matches what we expect based on MIME type to be extra safe
    # (Simplified for example purposes)
    safe_filename = f"{uuid.uuid4().hex}{ext}"
    
    save_path = os.path.join(UPLOAD_FOLDER, safe_filename)
    
    # 5. Save the file
    file.save(save_path)
    
    return "File uploaded securely!", 200
```
