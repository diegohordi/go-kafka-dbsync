# Kafka DB Sync

Study project that uses Apache Kafka as syncing mechanisms between a monolith DB and a microservice.

The main purpose of this project is to find out a solution to break up a monolith DB (usually a legacy DB) into 
a microservice, considering the fact that the monolith DB will still be used until further notice (for instance, 
development of other microservices, refactoring of integrations, and so on...).

Using [Sakila](https://dev.mysql.com/doc/index-other.html) sample as the monolith DB, I created a microservice that 
is responsible to deal only with the Catalogue context. This microservice is a REST API and, when users send out 
a new entry for the films catalogue, for instance, we should store this entry in our microservice DB and sync it 
with the monolith DB. The same is valid for the other way around, I mean, when the monolith DB is updated, we
should sync it with our microservice DB.

To accomplish that, as mentioned before, I used Apache Kafka as event stream platform to keep the loose coupling 
between the microservice and the legacy DBs. I also used a Kafka connector to receive continuously any changes in 
the legacy DB. For databases I used MySQL and the publishers, consumers and the Rest API were written in Go.

# Important notes
* For the sake of simplicity I didn't implement deletes, only updates and inserts, and the code was written as simple as possible;
* I created a UUID field in Film's monolith DB to keep some relation between the rows in both databases;
* The sync between the databases is [near real-time](https://www.kai-waehner.de/blog/2021/01/04/apache-kafka-is-not-hard-real-time-industrial-iot-embedded-connected-vehicles-automotive/);

# How to run
* `make run`
* `make create_source_connector`

# How to test
* Insert a film into catalogue: `curl -i -X POST http://localhost:8080/api/v1/catalogue -H "Content-Type: application/json" -d '{"title": "The Sixth Sense", "year": 1999}'`
* You can check if the event was correctly sent to Kafka, you can access http://localhost:9000 and check the catalogue topic
* In both databases you should be able to see the film created
* Perform some change in the Title or Year film's columns from legacy DB and check if these changes are sync: `curl -i -X GET http://localhost:8080/api/v1/catalogue/711a38b0-038a-49c9-a27c-f6780c2b649d`
* Update the film using REST API `curl -i -X PUT http://localhost:8080/api/v1/catalogue/711a38b0-038a-49c9-a27c-f6780c2b649d -H "Content-Type: application/json" -d '{"title": "The Sixth Sense", "year": 2021}'`
* Keep playing =)
