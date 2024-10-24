# Auto-generated by fogg. Do not edit
# Make improvements in fogg, so that everyone can benefit.
# tflint-ignore: terraform_unused_declarations
data "terraform_remote_state" "global" {
  backend = "s3"
  config = {


    bucket = "buck"

    key    = "terraform/proj/global.tfstate"
    region = "us-west-2"

    assume_role = {
      role_arn = "arn:aws:iam::123456789012:role/role"
    }

  }
}
# tflint-ignore: terraform_unused_declarations
data "terraform_remote_state" "lambda" {
  backend = "s3"
  config = {


    bucket = "buck"

    key    = "terraform/proj/envs/test/components/lambda.tfstate"
    region = "us-west-2"

    assume_role = {
      role_arn = "arn:aws:iam::123456789012:role/role"
    }

  }
}
# tflint-ignore: terraform_unused_declarations
data "terraform_remote_state" "network" {
  backend = "s3"
  config = {


    bucket = "buck"

    key    = "terraform/proj/envs/test/components/network.tfstate"
    region = "us-west-2"

    assume_role = {
      role_arn = "arn:aws:iam::123456789012:role/role"
    }

  }
}
