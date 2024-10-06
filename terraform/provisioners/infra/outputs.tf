output "ecr-repository-api_url" {
  value = module.ecr-repository-api.repository_url
}

output "s3-dictionary-bucket_name" {
  value = module.s3-dictionary-bucket.s3_name
}

output "s3-dictionary-bucket_arn" {
  value = module.s3-dictionary-bucket.s3_arn
}

output "s3-processing-bucket_name" {
  value = module.s3-processing-bucket.s3_name
}

output "s3-processing-bucket_arn" {
  value = module.s3-processing-bucket.s3_arn
}

output "dynamo-dictionary-table_name" {
  value = module.dynamo-dictionary-table.table_name
}

output "dynamo-dictionary-table_arn" {
  value = module.dynamo-dictionary-table.table_arn
}

output "dynamo-dictionary-stream_arn" {
  value = module.dynamo-dictionary-table.stream_arn
}

output "dynamo-logs-table_name" {
  value = module.dynamo-logs-table.table_name
}

output "dynamo-logs-table_arn" {
  value = module.dynamo-logs-table.table_arn
}

output "dynamo-logs-stream_arn" {
  value = module.dynamo-logs-table.stream_arn
}

output "sqs-put-csv-dead-letter-queue_url" {
  value = module.dictionary_put_csv_queue.dead_letter_queue_url
}

output "sqs-put-csv-dead-letter-queue_arn" {
  value = module.dictionary_put_csv_queue.dead_letter_queue_arn
}

output "sqs-put-csv-queue_url" {
  value = module.dictionary_put_csv_queue.queue_url
}

output "sqs-put-csv-queue_arn" {
  value = module.dictionary_put_csv_queue.queue_arn
}
