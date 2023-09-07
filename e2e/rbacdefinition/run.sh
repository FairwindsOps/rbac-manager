BASE_DIR=$(dirname $BASH_SOURCE)

printf "\n\n"
echo "********************************************************************"
echo "** Test rbacDefinition **"
echo "********************************************************************"
printf "\n\n"


bash "$BASE_DIR/cluterrolebindings/main.sh"
if [ $? -ne 0 ]; then
  exit 1
fi