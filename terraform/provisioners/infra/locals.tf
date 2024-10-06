locals {
  project      = "lingocards"
  state_bucket = "tfstates-${local.project}"
  tfstate_file = "infra.tfstates"

  dictionary_dynamo_schema = jsondecode(
    file("${path.module}/../../../dynamodb-interface/.tmpl/dynamo_dictionary_table.json")
  )

  logs_dynamo_schema = jsondecode(
    file("${path.module}/../../../dynamodb-interface/.tmpl/dynamo_logs_table.json")
  )
}