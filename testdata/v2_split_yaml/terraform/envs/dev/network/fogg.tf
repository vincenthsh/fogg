# Auto-generated by fogg. Do not edit
# Make improvements in fogg, so that everyone can benefit.

provider "assert" {}
terraform {
  required_version = "=1.1.1"

  backend "s3" {

    bucket = "bucket"

    key     = "terraform/foo/envs/dev/components/network.tfstate"
    encrypt = true
    region  = "region"
    profile = "foo"


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
  default = "dev"
}
# tflint-ignore: terraform_unused_declarations
variable "project" {
  type    = string
  default = "foo"
}
# tflint-ignore: terraform_unused_declarations
# tflint-ignore: terraform_unused_declarations
variable "component" {
  type    = string
  default = "network"
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
    project    = "foo"
    env        = "dev"
    service    = "network"
    owner      = "foo@example.com"
    tfstateKey = "terraform/foo/envs/dev/components/network.tfstate"

    managedBy = "terraform"
  }
}
# tflint-ignore: terraform_unused_declarations
variable "aws_accounts" {
  type = map(string)
  default = {

  }
}
