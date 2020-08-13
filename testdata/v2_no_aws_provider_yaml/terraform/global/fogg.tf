# Auto-generated by fogg. Do not edit
# Make improvements in fogg, so that everyone can benefit.
terraform {
  required_version = "~>0.100.0"

  backend s3 {

    bucket = "buck"

    key     = "terraform/proj/global.tfstate"
    encrypt = true
    region  = "us-west-2"
    profile = "profile"


  }
}
variable env {
  type    = string
  default = ""
}
variable project {
  type    = string
  default = "proj"
}
variable component {
  type    = string
  default = "global"
}
variable owner {
  type    = string
  default = "foo@example.com"
}
variable tags {
  type = object({ project : string, env : string, service : string, owner : string, managedBy : string })
  default = {
    project   = "proj"
    env       = ""
    service   = "global"
    owner     = "foo@example.com"
    managedBy = "terraform"
  }
}
variable foo {
  type    = string
  default = "bar1"
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
