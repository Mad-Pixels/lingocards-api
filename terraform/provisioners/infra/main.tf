module "ecr-repository-api" {
  source = "../../modules/ecr"

  project         = local.project
  repository_name = "images"
}

module "s3-dictionary-bucket" {
  source = "../../modules/s3"

  project     = local.project
  bucket_name = "dictionary"
}

module "s3-processing-bucket" {
  source = "../../modules/s3"

  project     = local.project
  bucket_name = "processing"
}

module "s3-errors-bucket" {
  source = "../../modules/s3"

  project     = local.project
  bucket_name = "errors"

  rule = {
    id     = "cleanup"
    status = "Enabled"
    filter = {
      prefix = "logs-"
    }
    expiration = {
      days = 30
    }
  }
}

module "dynamo-dictionary-table" {
  source = "../../modules/dynamo"

  project              = local.project
  table_name           = local.dictionary_dynamo_schema.table_name
  hash_key             = local.dictionary_dynamo_schema.hash_key
  range_key            = local.dictionary_dynamo_schema.range_key
  attributes           = local.dictionary_dynamo_schema.attributes
  secondary_index_list = local.dictionary_dynamo_schema.secondary_indexes
  stream_enabled       = true
}

module "dictionary_put_csv_queue" {
  source = "../../modules/sqs"

  project    = local.project
  queue_name = "put"
}