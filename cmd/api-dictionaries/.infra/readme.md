# Description

Lambda for manage dictionaries.

# Examples
## Define variables

```bash
api="xw66q4bfqv"

# localstack
url="http://localhost:4566/restapis/${api}/prod/_user_request_"

api_path_put="v1/dictionary/manage/put"
api_path_delete="v1/dictionary/manage/delete"
api_path_presign="v1/dictionary/manage/upload_url"
```

## v1/dictionary/manage/put
```bash
curl -X POST ${url}/${api_path_put} \
    -d '{"description": "description", "filename": "1.csv", "name": "test dictionary", "author": "author", "category": "language", "subcategory": "ru-he", "is_public": true}' \
    -H "Content-Type: application/json" 
```

## v1/dictionary/manage/upload_url
```bash
curl -X POST ${url}/${api_path_presign} \
  -d '{"content_type": "text/csv", "name": "1.csv"}' \
  -H "Content-Type: application/json"
```