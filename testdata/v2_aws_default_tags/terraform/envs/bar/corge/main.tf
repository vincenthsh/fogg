# Auto-generated by fogg. Do not edit
# Make improvements in fogg, so that everyone can benefit.

module "my_module" {
  source = "../../../modules/my_module"

  providers = {
    aws                 = aws
    aws.no_default_tags = aws.no_default_tags
  }
}
