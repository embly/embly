terraform {
  backend "s3" {
    bucket  = "embly-app-tfstate"
    key     = "embly-app"
    profile = "max"
    region  = "us-east-1"
  }
}

resource "aws_s3_bucket" "embly_app_tfstate" {
  bucket = "embly-app-tfstate"
  acl    = "private"

  versioning {
    enabled = true
  }
}

provider "aws" {
  region                  = "us-east-1"
  shared_credentials_file = "~/.aws/credentials"
  profile                 = "max"
  version                 = "~> 2.7"
}
