# Auto-generated by fogg. Do not edit
# Make improvements in fogg, so that everyone can benefit.
# tflint-ignore: terraform_unused_declarations
data "terraform_remote_state" "global" {
  backend = "s3"
  config = {


    bucket = "buck"

    key     = "terraform/proj/global.tfstate"
    region  = "us-west-2"
    profile = "profile"


  }
}
# tflint-ignore: terraform_unused_declarations
data "terraform_remote_state" "comp2" {
  backend = "remote"
  config = {


    hostname     = "example.com"
    organization = "foo"
    workspaces = {
      name = "staging-comp2"
    }

  }
}
# tflint-ignore: terraform_unused_declarations
data "terraform_remote_state" "vpc" {
  backend = "remote"
  config = {


    hostname     = "example.com"
    organization = "foo"
    workspaces = {
      name = "staging-vpc"
    }

  }
}
# tflint-ignore: terraform_unused_declarations
data "terraform_remote_state" "bar" {
  backend = "s3"
  config = {


    bucket = "buck"

    key     = "terraform/proj/accounts/bar.tfstate"
    region  = "us-west-2"
    profile = "profile"


  }
}
# tflint-ignore: terraform_unused_declarations
data "terraform_remote_state" "foo" {
  backend = "s3"
  config = {


    bucket = "buck"

    key     = "terraform/proj/accounts/foo.tfstate"
    region  = "us-west-2"
    profile = "profile"


  }
}
