# Example: Safe Subprocess Execution

This example demonstrates how to safely call external programs in Python and Node.js by avoiding the shell.

## ❌ Vulnerable Example (Python)

```python
import subprocess

def ping_host(user_provided_ip):
    # BAD: Using shell=True and string concatenation.
    # If user_provided_ip is "127.0.0.1; cat /etc/passwd", the shell will execute both commands.
    command = f"ping -c 4 {user_provided_ip}"
    result = subprocess.run(command, shell=True, capture_output=True, text=True)
    return result.stdout
```

## ✅ Secure Example (Python)

```python
import subprocess

def ping_host_secure(user_provided_ip):
    # GOOD: Using a list of arguments and shell=False (the default in subprocess.run)
    # If user_provided_ip is "127.0.0.1; cat /etc/passwd", the OS tries to ping a literal 
    # host named "127.0.0.1; cat /etc/passwd", which will safely fail.
    command = ["ping", "-c", "4", user_provided_ip]
    
    try:
        # shell=False is implied when passing a list
        result = subprocess.run(command, capture_output=True, text=True, check=True)
        return result.stdout
    except subprocess.CalledProcessError as e:
        return f"Ping failed: {e}"
```

## ❌ Vulnerable Example (Node.js)

```javascript
const { exec } = require('child_process');

function listFiles(userDir) {
    // BAD: exec() spawns a shell.
    exec(`ls -l ${userDir}`, (error, stdout, stderr) => {
        console.log(stdout);
    });
}
```

## ✅ Secure Example (Node.js)

```javascript
const { spawn, execFile } = require('child_process');

function listFilesSecure(userDir) {
    // GOOD: execFile() or spawn() does not spawn a shell by default.
    // It passes arguments safely.
    execFile('ls', ['-l', userDir], (error, stdout, stderr) => {
        console.log(stdout);
    });
}
```
