package config

var exampleConfig = `
#tf-plan-reporter tool example config file
terraform_binary_file: /usr/bin/terraform     # Absolute or relative path of terraform command. MANDATORY parameter
terraform_plan_file_basename: plan.bin        # Base name of terraform binary file for further search. MANDATORY parameter
terraform_plan_search_folder: .               # Common parent folder from which to start search of generated plan files. MANDATORY parameter

# List of cloud resources which should be kept. Must be either "all" and then the "allowed_removals" section of this file
# is going to payed attention. Or it should contain particular type of resources which should be kept from accidental removals. OPTIONAL parameter
critical_resources:
  - all

allowed_removals:                             # Makes sense only if "all" item is specified in "critical_resources" section. OPTIONAL parameter
  - null_resource
  - azurerm_role_assignment
  - azurerm_monitor_diagnostic_setting
  - azurerm_key_vault

not_use_chdir: false                          # Whether of not the arg '-chdir=DIR' of 'terraform' command is going be used for processing of TF binary file.
                                              # In short, if you're using terraform+terragrant bunch for cloud provisioning this parameter should be 'false'.
                                              # If you're using terraform only, it needs to be set up to 'true'
`

func PrintExample() {
    println(exampleConfig)
}
