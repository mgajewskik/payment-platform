provider "aws" {
  region = var.aws_region
}

locals {
  resource_prefix = "payment-platform"
}

resource "aws_dynamodb_table" "this" {
  name         = "${local.resource_prefix}-table"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "PK"
  range_key    = "SK"

  attribute {
    name = "PK"
    type = "S"
  }

  attribute {
    name = "SK"
    type = "S"
  }

  server_side_encryption {
    enabled = true
  }
}
