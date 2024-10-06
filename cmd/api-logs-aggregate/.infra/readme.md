# Description

Lambda for aggregate logs from devices.

# Examples
## Define variables

```bash
api="zlaejkxrcb"

# localstack
url="http://localhost:4566/restapis/${api}/prod/_user_request_"
token="000XXX000"

device_path_put="device/v1/logs/put"
```

## Define body
```bash
body='{"timestamp":1234345,"os_version":"18.0","device":"iphone","error_type": "error","app_version":"0.0.0","additional_info":"info","error_message": "Failed to connect to server", "additional_info":"asd"}'
```

## /device/v1/logs/put
```bash
timestamp=$(date -u +%s)
signature=$(echo -n "${timestamp}" | openssl dgst -sha256 -hmac "${token}" | sed 's/^.* //')
curl -X POST ${url}/${device_path_put} \
    -d "${body}" \
    -H "Content-Type: application/json" \
    -H "x-timestamp: ${timestamp}" \
    -H "x-signature: ${signature}"
```
