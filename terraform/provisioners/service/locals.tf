locals {
  _root_dir = "${path.module}/../../../cmd"
  _entries  = fileset(local._root_dir, "**")

  _lambda_functions = distinct([
    for v in local._entries : split("/", v)[0]
    if length(split("/", v)) > 1
  ])

  _lambda_configs = {
    for func in local._lambda_functions :
    func => fileexists("${local._root_dir}/${func}/.infra/config.json") ?
    jsondecode(
      templatefile("${local._root_dir}/${func}/.infra/config.json", local.template_vars)
    ) : null
  }

  lambdas      = { for func in local._lambda_functions : func => local._lambda_configs[func] }
  project      = "lingocards"
  state_bucket = "tfstates-madpixels"
  tfstate_file = "lingocards-api/infra.tfstate"

  // template variables which use in ./infra/config.json of each lambda.
  template_vars = {
    var_device_api_token        = var.device_api_token
    dictionary_bucket_name      = data.terraform_remote_state.infra.outputs.s3-dictionary-bucket_name
    processing_bucket_name      = data.terraform_remote_state.infra.outputs.s3-processing-bucket_name
    dictionary_bucket_arn       = data.terraform_remote_state.infra.outputs.s3-dictionary-bucket_arn
    processing_bucket_arn       = data.terraform_remote_state.infra.outputs.s3-processing-bucket_arn
    dictionary_table_arn        = data.terraform_remote_state.infra.outputs.dynamo-dictionary-table_arn
    dictionary_table_stream_arn = data.terraform_remote_state.infra.outputs.dynamo-dictionary-stream_arn
    logs_table_arn              = data.terraform_remote_state.infra.outputs.dynamo-logs-table_arn
    logs_table_stream_arn       = data.terraform_remote_state.infra.outputs.dynamo-logs-stream_arn
    put_csv_sqs_queue_url       = data.terraform_remote_state.infra.outputs.sqs-put-csv-queue_url
    put_csv_sqs_queue_arn       = data.terraform_remote_state.infra.outputs.sqs-put-csv-queue_arn
  }
}
