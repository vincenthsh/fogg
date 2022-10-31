# Auto-generated by fogg. Do not edit
# Make improvements in fogg, so that everyone can benefit.
provider "aws" {

  region = "us-west-2"


  assume_role {
    role_arn = "arn:aws:iam::626314663666:role/tfe-si"
  }

  allowed_account_ids = ["626314663666"]
}
# Aliased Providers (for doing things in every region).


provider "tfe" {
  hostname = "si.prod.tfe.czi.technology"
}
terraform {
  required_version = "=1.2.6"

  backend "remote" {

    hostname     = "si.prod.tfe.czi.technology"
    organization = "shared-infra"
    workspaces {
      name = "tfe"
    }

  }
  required_providers {

    archive = {
      source = "hashicorp/archive"

      version = "~> 2.0"

    }

    assert = {
      source = "bwoznicki/assert"

      version = "~> 0.0.1"

    }

    aws = {
      source = "hashicorp/aws"

      version = "3.30.0"

    }

    local = {
      source = "hashicorp/local"

      version = "~> 2.0"

    }

    null = {
      source = "hashicorp/null"

      version = "~> 3.0"

    }

    okta-head = {
      source = "okta/okta"

      version = "~> 3.30"

    }

    random = {
      source = "hashicorp/random"

      version = "~> 3.4"

    }

    tfe = {
      source = "hashicorp/tfe"

      version = "0.33.0"

    }

    tls = {
      source = "hashicorp/tls"

      version = "~> 3.0"

    }

  }
}
variable "env" {
  type    = string
  default = ""
}
variable "project" {
  type    = string
  default = "foo"
}
variable "region" {
  type    = string
  default = "us-west-2"
}
variable "component" {
  type    = string
  default = ""
}
variable "owner" {
  type    = string
  default = "foo@example.com"
}
variable "tags" {
  type = object({ project : string, env : string, service : string, owner : string, managedBy : string })
  default = {
    project   = "foo"
    env       = ""
    service   = ""
    owner     = "foo@example.com"
    managedBy = "terraform"
  }
}
variable "TFE_AWS_ACCESS_KEY_ID" {
  type    = string
  default = ""
}
variable "TFE_AWS_SECRET_ACCESS_KEY" {
  type    = string
  default = ""
}
variable "aws_accounts" {
  type = map(string)
  default = {

  }
}