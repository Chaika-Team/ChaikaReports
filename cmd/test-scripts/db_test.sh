#!/bin/bash

# Variables
CASSANDRA_HOST="188.242.205.5"    # Your Cassandra host
CASSANDRA_USER=""           # Your Cassandra username
CASSANDRA_PASS=""  # Your Cassandra password
MAIN_KEYSPACE="chaika_sales"           # Main keyspace name
TEST_KEYSPACE="chaika_sales_test"      # Test keyspace name
TABLE_NAME="operations"                # Table name to copy
CSV_FILE="operations_data.csv"         # Name of the CSV file
CSV_DIR="tmp_csv"                      # Temporary directory for CSV files
SAMPLE_SIZE=1000000                    # Number of rows to fetch for testing

# Step 1: Create the table in the test keyspace
echo "Creating table in the test keyspace..."

# Get the CREATE TABLE statement from the main keyspace and write it to a file
cqlsh $CASSANDRA_HOST \
     -u $CASSANDRA_USER -p "$CASSANDRA_PASS" --no-color \
     -e "DESCRIBE TABLE $MAIN_KEYSPACE.$TABLE_NAME;" \
     | sed -n '/^CREATE TABLE/, $p' > create_table.cql

# Replace the keyspace name in the CREATE TABLE statement
sed -i "s/CREATE TABLE $MAIN_KEYSPACE\./CREATE TABLE $TEST_KEYSPACE./g" create_table.cql

# Execute the CREATE TABLE statement in the test keyspace
cqlsh $CASSANDRA_HOST \
     -u $CASSANDRA_USER -p "$CASSANDRA_PASS" \
     -f create_table.cql

# Check if the previous command was successful
if [ $? -ne 0 ]; then
  echo "Failed to create table in the test keyspace."
  read -p "Press any key to exit"
  exit 1
fi

# Remove the create_table.cql file
rm create_table.cql

# Step 2: Copy data using DSBulk with a limit of 1,000,000 entries
echo "Copying a subset of data ($SAMPLE_SIZE entries) from the main table to the test table using DSBulk..."

# Create and clean the temporary CSV directory
mkdir -p $CSV_DIR
rm -rf $CSV_DIR/*

# Export data from the main table to the CSV directory
dsbulk unload \
  -h $CASSANDRA_HOST \
  --driver.hosts $CASSANDRA_HOST \
  --driver.resolveContactPoints false \
  -u $CASSANDRA_USER \
  -p "$CASSANDRA_PASS" \
  --driver.auth.allowPlaintextWithoutChallenge true \
  -k $MAIN_KEYSPACE \
  -t $TABLE_NAME \
  -url $CSV_DIR \
  --connector.csv.fileName $CSV_FILE \
  --connector.csv.maxConcurrentFiles 1 \
  --max-rows $SAMPLE_SIZE \
  -header true

# Check if the unload was successful
if [ $? -ne 0 ]; then
  echo "Data unload failed."
  read -p "Press any key to exit"
  exit 1
fi

# Import data from the CSV directory into the test table
dsbulk load \
  -h $CASSANDRA_HOST \
  --driver.hosts $CASSANDRA_HOST \
  --driver.resolveContactPoints false \
  -u $CASSANDRA_USER \
  -p "$CASSANDRA_PASS" \
  --driver.auth.allowPlaintextWithoutChallenge true \
  -k $TEST_KEYSPACE \
  -t $TABLE_NAME \
  -url $CSV_DIR \
  --connector.csv.fileName $CSV_FILE \
  --connector.csv.maxConcurrentFiles 1 \
  -header true

# Check if the load was successful
if [ $? -ne 0 ]; then
  echo "Data load failed."
  read -p "Press any key to exit"
  exit 1
fi

# Remove the CSV file and temporary directory
rm -rf $CSV_DIR

echo "Script completed successfully."

read -p "Press any key to exit"
