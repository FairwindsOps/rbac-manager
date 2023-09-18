BASE_DIR="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

printf "\n\n"
echo "********************************************************************"
echo "** Test rbacDefinition **"
echo "********************************************************************"
printf "\n\n"


bash "$BASE_DIR/cluterrolebindings/main.sh"
if [ $? -ne 0 ]; then
  exit 1
fi

bash "$BASE_DIR/serviceaccounts/main.sh"
if [ $? -ne 0 ]; then
  exit 1
fi