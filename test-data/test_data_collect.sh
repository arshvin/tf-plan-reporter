plans=(`find . -name tfplan.bin`)
current_dir=`pwd`
for plan in $plans; do
module_dir=`dirname $plan`
module_name=`basename $module_dir`

cd "$current_dir/$module_dir"
terraform show -no-color -json tfplan.bin > "$current_dir/$module_name/plan.json"
done
