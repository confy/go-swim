resource "aws_cloudwatch_event_rule" "daily" {
  name       = "${local.function_name}-daily"
  description = "run daily at 9am"
  schedule_expression = "cron(0 16 * * ? *)" # 4pm UTC (9am PST)
}

resource "aws_cloudwatch_event_target" "go-swim-target" {
  arn = aws_lambda_function.function.arn
  rule = aws_cloudwatch_event_rule.daily.name
}

resource "aws_lambda_permission" "allow_cloudwatch_to_call_function" {
  statement_id = "AllowExecutionFromCloudWatch"
  action = "lambda:InvokeFunction"
  function_name = aws_lambda_function.function.function_name
  principal = "events.amazonaws.com"
  source_arn = aws_cloudwatch_event_rule.daily.arn
}