# Example: Retrieving Secrets at Runtime (AWS Secrets Manager)

This Python example demonstrates how an application should retrieve a database password from AWS Secrets Manager at startup, rather than reading it from a file or hardcoded string.

**Prerequisite**: The EC2 instance or ECS/EKS task running this code must have an IAM Role attached that grants `secretsmanager:GetSecretValue` permission for the specific secret ARN.

```python
import boto3
import json
from botocore.exceptions import ClientError

def get_db_credentials():
    secret_name = "prod/MyService/DatabaseCredentials"
    region_name = "us-east-1"

    # Create a Secrets Manager client
    # Boto3 automatically picks up the IAM role credentials from the environment
    session = boto3.session.Session()
    client = session.client(
        service_name='secretsmanager',
        region_name=region_name
    )

    try:
        get_secret_value_response = client.get_secret_value(
            SecretId=secret_name
        )
    except ClientError as e:
        # Handled exceptions
        if e.response['Error']['Code'] == 'ResourceNotFoundException':
            print("The requested secret was not found")
        elif e.response['Error']['Code'] == 'InvalidRequestException':
            print("The request was invalid due to:")
        elif e.response['Error']['Code'] == 'InvalidParameterException':
            print("The request had invalid params:")
        elif e.response['Error']['Code'] == 'DecryptionFailure':
            print("The requested secret can't be decrypted using the provided KMS key:")
        elif e.response['Error']['Code'] == 'InternalServiceError':
            print("An error occurred on service side:")
        raise e
    else:
        # Secrets Manager decrypts the secret value using the associated KMS CMK
        if 'SecretString' in get_secret_value_response:
            secret = get_secret_value_response['SecretString']
            
            # The secret is typically stored as a JSON string
            credentials = json.loads(secret)
            return credentials['username'], credentials['password']
        else:
            # Handle binary secrets
            decoded_binary_secret = base64.b64decode(get_secret_value_response['SecretBinary'])
            return None, decoded_binary_secret

# Usage in application startup
db_user, db_pass = get_db_credentials()

# Connect to database using db_user and db_pass ...
# WARNING: Never log db_pass!
```
