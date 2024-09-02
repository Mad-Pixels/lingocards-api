output "function_arn" {
  description = "ARN of the Lambda function"
  value       = aws_lambda_function.container_function.arn
}

output "function_name" {
  description = "Name of the Lambda function"
  value       = aws_lambda_function.container_function.function_name
}