data "terraform_remote_state" "dep" {
  backend = "http"

  config = {
    address = "https://grid.example.com/tfstate/11111111-1111-1111-1111-111111111111"
  }
}

output "example" {
  value = data.terraform_remote_state.dep.outputs.found_output
}
