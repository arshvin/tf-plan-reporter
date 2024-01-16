terragrunt plan-all --terragrunt-working-dir root_of_modules/ --terragrunt-non-interactive -out=tfplan.bin -lock=false


########
plans=(`find ../ -name tfplan.bin`)
current_dir=`pwd`
for plan in $plans; do
module_dir=`dirname $plan`
module_name=`basename $module_dir`

mkdir "$current_dir/$module_name"
echo "Processing of $current_dir/$module_name"
terraform -chdir="$current_dir/$module_dir" show -no-color -json tfplan.bin > "$current_dir/$module_name/plan.json"
done
