terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
    }
    archive = {
      source = "hashicorp/archive"
    }
    null = {
      source = "hashicorp/null"
    }
  }

  required_version = ">= 1.3.3"
}

provider "aws" {
  region  = "us-west-2"

  default_tags {
    tags = {
      app = "go-swim"
    }
  }
}