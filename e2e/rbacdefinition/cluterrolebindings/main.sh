BASE_DIR="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

printf "\n\n"
echo "********************************************************************"
echo "** Test clusterrolebindings **"
echo "********************************************************************"
printf "\n\n"

# Execute the setup, then execute the tests just if the setup contains no errors.
# Finally always execute the cleanup and return the whole error of the steps
error=$((0))
bash "$BASE_DIR/setup.sh"
error=$(( error | $? ))

if [ $error -eq 0 ]; then
bash "$BASE_DIR/tests.sh"
error=$(( error | $? ))
fi

bash "$BASE_DIR/cleanup.sh"
exit $(( error | $? ))
