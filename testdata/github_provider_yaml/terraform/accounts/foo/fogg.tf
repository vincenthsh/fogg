# Auto-generated by fogg. Do not edit
# Make improvements in fogg, so that everyone can benefit.

provider github {
  organization = "foo"
  base_url     = "https://example.com/"
}
terraform {
  required_version = "=1.1.1"
  backend s3 {

    bucket = "bucket"

    key     = "terraform/foo/accounts/foo.tfstate"
    encrypt = true
    region  = "region"
    profile = "foo"

  }
}
variable project {
  type    = string
  default = "foo"
}
variable owner {
  type    = string
  default = "foo@example.com"
}
# map of aws_accounts
variable aws_accounts {
  type = map
  default = {



  }
}
data terraform_remote_state global {
  backend = "s3"
  config = {


    bucket = "bucket"

    key     = "terraform/foo/global.tfstate"
    region  = "region"
    profile = "foo"

  }
}
provider random {
  version = "~> 2.2"
}
provider template {
  version = "~> 2.1"
}
provider archive {
  version = "~> 1.3"
}
provider null {
  version = "~> 2.1"
}
provider local {
  version = "~> 1.4"
}
provider tls {
  version = "~> 2.1"
}
