# Auto-generated by fogg. Do not edit
# Make improvements in fogg, so that everyone can benefit.

module "network" {
  source          = "terraform-aws-modules/vpc/aws"
  version         = "5.1.2"
  azs             = local.azs
  cidr            = local.cidr
  name            = local.name
  private_subnets = local.private_subnets
  public_subnets  = local.public_subnets
  tags            = local.tags
}

module "my_module" {
  source = "../../../modules/my_module"
}
