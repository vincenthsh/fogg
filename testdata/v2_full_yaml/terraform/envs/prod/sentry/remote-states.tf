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
data "terraform_remote_state" "hero" {
  backend = "s3"
  config = {


    bucket = "buck"

    key     = "terraform/proj/envs/prod/components/hero.tfstate"
    region  = "us-west-2"
    profile = "profile"


  }
}
