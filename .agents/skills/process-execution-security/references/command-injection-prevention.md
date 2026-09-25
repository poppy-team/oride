# Command Injection Prevention

Command Injection (or OS Command Injection) occurs when an application passes unsafe user-supplied data to a system shell. In this attack, the attacker-supplied operating system commands are usually executed with the privileges of the vulnerable application.

## Vulnerability Mechanism

When executing external programs (like calling `ping`, `ls`, or ImageMagick tools), languages often provide a way to pass a string to a shell (e.g., `bash`, `cmd.exe`). If user input is concatenated into this string, an attacker can append their own commands.

**Attack payload example**: `; rm -rf /` or `&& cat /etc/passwd`

## Primary Defense: Avoid Shell Execution

The most effective way to prevent command injection is to **never invoke a system shell**.

Instead of passing a single string to a shell, pass an **array of arguments** directly to the operating system's execution API (`execve` on Unix). This API treats the first argument as the executable path and subsequent arguments strictly as parameters, preventing them from being interpreted as secondary commands or shell metacharacters.

## Secondary Defense: Validation

If you must execute a command based on user input, strictly validate the input:
- **Positive Allowlist**: Only allow specific, known-good alphanumeric characters.
- **Map to Enums**: Instead of taking arbitrary string input, have the user select an ID (1, 2, 3), and map that ID to the safe command string on the backend.

## Dangerous Functions to Avoid

- **Python**: `os.system()`, `os.popen()`, `subprocess.call(..., shell=True)`
- **Node.js**: `child_process.exec()`
- **PHP**: `exec()`, `shell_exec()`, `system()`, `passthru()`, `popen()`
- **Ruby**: `` `backticks` ``, `system()`, `exec()` (when passed a single string)
- **C/C++**: `system()`, `popen()`
