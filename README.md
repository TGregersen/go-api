# go-api
A repository for an implementation of a receipt processing api. 

To install:

Create a directory and include the following:

dockerfile
go.mod
go.sum
receipts.go

Once that folder has been established open a command window and navigate to the directory where the files are located.

Run the following command in that directory to build the docker container:

"docker build --tag receipt ."

finally start the container via docker desktop or the following command:

"docker run receipt"


Perform POST tests using the JSON template file.