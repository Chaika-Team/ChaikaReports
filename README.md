# ChaikaReports
**ChaikaReports** is a service that takes reports sent from ChaikaKotlin, processes the sent data and stores it in our Cassandra database. </br>
**ChaikaReports** is also capable of taking information from the database (Carts an employee has created during a trip, trips an employe has been in in a year, etc.) and sending to the person of interest. </br>

## ChaikaReports Technology:
+ Go
+ Cassandra
+ gocql

## ChaikaReports' main functionality consists of:
+ InsertData (Insert a received report into the database)
+ GetEmployeeCartsInTrip (Get the carts that were created by an employee during a trip)
+ GetEmployeeIDsByTrip (Get the ID's of the employees present during a trip)
+ GetEmployeeTrips (Get the trips an employee was in during a certain year)
+ UpdateItemQuantity (Update the quantity of items that were sold in a transaction)
+ DeleteItemFromCart (Delete an item in a cart)
