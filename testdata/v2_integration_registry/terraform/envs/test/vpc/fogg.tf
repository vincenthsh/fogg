# Auto-generated by fogg. Do not edit
# Make improvements in fogg, so that everyone can benefit.
provider "aws" {

  region  = "us-west-2"
  profile = "profile"

  allowed_account_ids = ["00456"]
}
# Aliased Providers (for doing things in every region).


provider "aws" {
  alias   = "stg"
  region  = "us-west-2"
  profile = "stg"

  allowed_account_ids = ["00456"]
}


provider "aws" {
  alias   = "prd"
  region  = "us-west-2"
  profile = "prd"

  allowed_account_ids = ["00456"]
}


provider "aws" {
  alias   = "no_default_tags"
  region  = "us-west-2"
  profile = "profile"

  allowed_account_ids = ["00456"]
}


provider "assert" {}
terraform {
  required_version = "=0.100.0"

  backend "s3" {

    bucket = "buck"

    key     = "terraform/proj/envs/test/components/vpc.tfstate"
    encrypt = true
    region  = "us-west-2"
    profile = "profile"


  }
  required_providers {
    archive = {
      source  = "hashicorp/archive"
      version = "~> 2.0"
    }
    assert = {
      source  = "bwoznicki/assert"
      version = "0.0.1"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "0.12.0"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.0"
    }
    null = {
      source  = "hashicorp/null"
      version = "~> 3.0"
    }
    okta-head = {
      source  = "okta/okta"
      version = "~> 3.30"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.4"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 3.0"
    }
  }
}
# tflint-ignore: terraform_unused_declarations
variable "env" {
  type    = string
  default = "test"
}
# tflint-ignore: terraform_unused_declarations
variable "project" {
  type    = string
  default = "proj"
}
# tflint-ignore: terraform_unused_declarations
variable "region" {
  type    = string
  default = "us-west-2"
}
# tflint-ignore: terraform_unused_declarations
variable "component" {
  type    = string
  default = "vpc"
}
variable "aws_profile" {
  type    = string
  default = "profile"
}
# tflint-ignore: terraform_unused_declarations
variable "owner" {
  type    = string
  default = "foo@example.com"
}
# tflint-ignore: terraform_unused_declarations
variable "tags" {
  type = object({ project : string, env : string, service : string, owner : string, managedBy : string, tfstateKey : string })
  default = {
    project    = "proj"
    env        = "test"
    service    = "vpc"
    owner      = "foo@example.com"
    tfstateKey = "terraform/proj/envs/test/components/vpc.tfstate"

    managedBy = "terraform"
  }
}
variable "foo" {
  type    = string
  default = "bar1"
}
# tflint-ignore: terraform_unused_declarations
variable "aws_accounts" {
  type = map(string)
  default = {

  }
}
